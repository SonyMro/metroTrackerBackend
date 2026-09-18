package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"metry.uk/provext4/MetroTrackerBackend/internal/users/domain"
)

type UserRepository struct{ db *pgxpool.Pool }

func NewUserRepository(db *pgxpool.Pool) *UserRepository { return &UserRepository{db: db} }

func (r *UserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	const query = `INSERT INTO usuario (nombre, apepat, apemat, foto, correo, password, fecharegistro, fechaact, idrol, idgenero, esactivo)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW(), $7, $8, TRUE)
		RETURNING id, fecharegistro, fechaact, esactivo`
	err := r.db.QueryRow(ctx, query, user.Nombre, user.Apepat, user.Apemat, user.Foto, user.Correo, user.Password, user.IDRol, user.IDGenero).Scan(&user.ID, &user.FechaRegistro, &user.FechaAct, &user.EsActivo)
	if err != nil {
		return nil, fmt.Errorf("insertar usuario: %w", err)
	}
	return user, nil
}

const userDetailColumns = `u.id, u.nombre, u.apepat, u.apemat, u.foto, u.correo, u.password, u.fecharegistro, u.fechaact, u.idrol, u.idgenero, u.esactivo, r.id, r.nombre, g.id, g.Tipo`
const userListColumns = `u.id, u.nombre, u.apepat, u.apemat, u.correo, u.fecharegistro, u.fechaact, u.idrol, u.idgenero, u.esactivo, r.id, r.nombre, g.id, g.Tipo`
const userJoins = ` FROM usuario u LEFT JOIN roles r ON r.id = u.idrol LEFT JOIN genero g ON g.id = u.idgenero `

func scanUser(row pgx.Row) (*domain.User, error) {
	user := &domain.User{Rol: &domain.Role{}, Genero: &domain.Genero{}}
	err := row.Scan(&user.ID, &user.Nombre, &user.Apepat, &user.Apemat, &user.Foto, &user.Correo, &user.Password, &user.FechaRegistro, &user.FechaAct, &user.IDRol, &user.IDGenero, &user.EsActivo, &user.Rol.ID, &user.Rol.Nombre, &user.Genero.ID, &user.Genero.Tipo)
	return user, err
}

func scanListUser(row pgx.Row) (*domain.User, error) {
	user := &domain.User{Rol: &domain.Role{}, Genero: &domain.Genero{}}
	err := row.Scan(&user.ID, &user.Nombre, &user.Apepat, &user.Apemat, &user.Correo, &user.FechaRegistro, &user.FechaAct, &user.IDRol, &user.IDGenero, &user.EsActivo, &user.Rol.ID, &user.Rol.Nombre, &user.Genero.ID, &user.Genero.Tipo)
	return user, err
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
	const query = `SELECT id, nombre, apepat, apemat, foto, correo, password, fecharegistro, fechaact, idrol, idgenero, esactivo FROM usuario WHERE correo = $1`
	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(&user.ID, &user.Nombre, &user.Apepat, &user.Apemat, &user.Foto, &user.Correo, &user.Password, &user.FechaRegistro, &user.FechaAct, &user.IDRol, &user.IDGenero, &user.EsActivo)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("consultar usuario por correo: %w", err)
	}
	return user, nil
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	rows, err := r.db.Query(ctx, `SELECT `+userListColumns+userJoins+` ORDER BY u.id DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("consultar lista de usuarios: %w", err)
	}
	defer rows.Close()
	users := make([]*domain.User, 0)
	for rows.Next() {
		user, err := scanListUser(rows)
		if err != nil {
			return nil, fmt.Errorf("leer usuario listado: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterar lista de usuarios: %w", err)
	}
	return users, nil
}

func (r *UserRepository) Count(ctx context.Context) (int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM usuario`).Scan(&total); err != nil {
		return 0, fmt.Errorf("contar usuarios: %w", err)
	}
	return total, nil
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
	result, err := r.db.Exec(ctx, `UPDATE usuario SET nombre = $1, apepat = $2, apemat = $3, idrol = $4, idgenero = $5, fechaact = NOW() WHERE id = $6`, user.Nombre, user.Apepat, user.Apemat, user.IDRol, user.IDGenero, user.ID)
	if err != nil {
		return fmt.Errorf("actualizar usuario: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}
