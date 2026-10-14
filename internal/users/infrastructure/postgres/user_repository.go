package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"metry.uk/provext4/MetroTrackerBackend/internal/users/domain"
)

type UserRepository struct{ db *pgxpool.Pool }

func NewUserRepository(db *pgxpool.Pool) *UserRepository { return &UserRepository{db: db} }

func (r *UserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	const query = `INSERT INTO usuario (nombre, apepat, apemat, foto, correo, password, fecharegistro, fechaact, idgenero, esactivo)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW(), $7, TRUE)
		RETURNING id, fecharegistro, fechaact, esactivo`
	err := r.db.QueryRow(ctx, query, user.Nombre, user.Apepat, user.Apemat, user.Foto, user.Correo, user.Password, user.IDGenero).Scan(&user.ID, &user.FechaRegistro, &user.FechaAct, &user.EsActivo)
	if err != nil {
		return nil, fmt.Errorf("insertar usuario: %w", err)
	}
	return user, nil
}

// ResolveGoogleIdentity mantiene la búsqueda, vinculación y alta en una sola transacción.
func (r *UserRepository) ResolveGoogleIdentity(ctx context.Context, claims domain.GoogleClaims, defaultGenderID int32) (*domain.User, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("iniciar transacción de identidad Google: %w", err)
	}
	defer tx.Rollback(ctx)

	user, err := r.resolveGoogleLinkedUser(ctx, tx, claims)
	if err == nil {
		if err := finalizeGoogleTransaction(ctx, tx, "confirmar identidad Google"); err != nil {
			return nil, err
		}
		return user, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("buscar vínculo Google: %w", err)
	}

	user, err = r.resolveGoogleIdentityByEmail(ctx, tx, claims)
	if err == nil {
		if err := finalizeGoogleTransaction(ctx, tx, "confirmar vínculo Google"); err != nil {
			return nil, err
		}
		return user, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("buscar usuario por correo en transacción: %w", err)
	}

	user, err = r.createGoogleUser(ctx, tx, claims, defaultGenderID)
	if err != nil {
		return nil, err
	}
	if err := finalizeGoogleTransaction(ctx, tx, "confirmar alta Google"); err != nil {
		return nil, err
	}
	return user, nil
}

func finalizeGoogleTransaction(ctx context.Context, tx pgx.Tx, action string) error {
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: %w", action, err)
	}
	return nil
}

func (r *UserRepository) resolveGoogleLinkedUser(ctx context.Context, tx pgx.Tx, claims domain.GoogleClaims) (*domain.User, error) {
	const linkedQuery = `SELECT ` + userDetailColumns + userJoins + ` JOIN user_auth a ON a.user_id = u.id WHERE a.provider = 'google' AND a.provider_id = $1 FOR UPDATE OF u`
	return scanUser(tx.QueryRow(ctx, linkedQuery, claims.Subject))
}

func (r *UserRepository) resolveGoogleIdentityByEmail(ctx context.Context, tx pgx.Tx, claims domain.GoogleClaims) (*domain.User, error) {
	const emailQuery = `SELECT ` + userDetailColumns + userJoins + ` WHERE u.correo = $1 FOR UPDATE OF u`
	user, err := scanUser(tx.QueryRow(ctx, emailQuery, claims.Email))
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `UPDATE usuario SET google_id = $1, fechaact = NOW() WHERE id = $2`, claims.Subject, user.ID); err != nil {
		return nil, fmt.Errorf("guardar vínculo Google en usuario: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO user_auth (user_id, provider, provider_id) VALUES ($1, 'google', $2)`, user.ID, claims.Subject); err != nil {
		return nil, googleLinkError(err)
	}
	return user, nil
}

func (r *UserRepository) createGoogleUser(ctx context.Context, tx pgx.Tx, claims domain.GoogleClaims, defaultGenderID int32) (*domain.User, error) {
	name := claims.Name
	if name == "" {
		name = claims.Email
	}
	var userID int32
	const insertUser = `INSERT INTO usuario (nombre, apepat, correo, google_id, fecharegistro, fechaact, idgenero, esactivo) VALUES ($1, '-', $2, $3, NOW(), NOW(), $4, TRUE) RETURNING id`
	if err := tx.QueryRow(ctx, insertUser, name, claims.Email, claims.Subject, defaultGenderID).Scan(&userID); err != nil {
		return nil, googleLinkError(err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO user_auth (user_id, provider, provider_id) VALUES ($1, 'google', $2)`, userID, claims.Subject); err != nil {
		return nil, googleLinkError(err)
	}
	user, err := scanUser(tx.QueryRow(ctx, `SELECT `+userDetailColumns+userJoins+` WHERE u.id = $1`, userID))
	if err != nil {
		return nil, fmt.Errorf("recuperar usuario Google: %w", err)
	}
	return user, nil
}

func googleLinkError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return fmt.Errorf("%w: %v", domain.ErrGoogleLinkAlreadyExists, err)
	}
	return fmt.Errorf("persistir identidad Google: %w", err)
}

const userDetailColumns = `u.id, u.nombre, u.apepat, COALESCE(u.apemat,''), u.foto, u.correo, COALESCE(u.password, ''), u.fecharegistro, u.fechaact, u.idgenero, u.esactivo, g.id, g.Tipo`
const userJoins = ` FROM usuario u LEFT JOIN genero g ON g.id = u.idgenero `

func scanUser(row pgx.Row) (*domain.User, error) {
	user := &domain.User{Genero: &domain.Genero{}}
	err := row.Scan(&user.ID, &user.Nombre, &user.Apepat, &user.Apemat, &user.Foto, &user.Correo, &user.Password, &user.FechaRegistro, &user.FechaAct, &user.IDGenero, &user.EsActivo, &user.Genero.ID, &user.Genero.Tipo)
	return user, err
}

func scanListGender(row pgx.Row) (*domain.Genero, error) {
	genero := &domain.Genero{}
	err := row.Scan(&genero.ID, &genero.Tipo)
	return genero, err
}

func (r *UserRepository) GetByID(ctx context.Context, id int32) (*domain.User, error) {
	user, err := scanUser(r.db.QueryRow(ctx, `SELECT `+userDetailColumns+userJoins+` WHERE u.id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("consultar usuario por ID: %w", err)
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const query = `SELECT id, nombre, apepat, apemat, foto, correo, password, fecharegistro, fechaact, idgenero, esactivo FROM usuario WHERE correo = $1`
	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(&user.ID, &user.Nombre, &user.Apepat, &user.Apemat, &user.Foto, &user.Correo, &user.Password, &user.FechaRegistro, &user.FechaAct, &user.IDGenero, &user.EsActivo)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("consultar usuario por correo: %w", err)
	}
	return user, nil
}

func (r *UserRepository) Deactivate(ctx context.Context, id int32) error {
	result, err := r.db.Exec(ctx, `UPDATE usuario SET esactivo = FALSE, fechaact = NOW() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("desactivar usuario: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	result, err := r.db.Exec(ctx, `UPDATE usuario SET nombre = $1, apepat = $2, apemat = $3, idgenero = $4, fechaact = NOW() WHERE id = $5`, user.Nombre, user.Apepat, user.Apemat, user.IDGenero, user.ID)
	if err != nil {
		return fmt.Errorf("actualizar usuario: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) ListGeneros(ctx context.Context) ([]*domain.Genero, error) {
	rows, err := r.db.Query(ctx, `select id, Tipo from genero`)
	if err != nil {
		return nil, fmt.Errorf("consultar lista de usuarios: %w", err)
	}
	defer rows.Close()
	generos := make([]*domain.Genero, 0)
	for rows.Next() {
		genero, err := scanListGender(rows)
		if err != nil {
			return nil, fmt.Errorf("leer géneros listados: %w", err)
		}
		generos = append(generos, genero)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterrar lista de generos: %w", err)

	}
	return generos, nil

}
