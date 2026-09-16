package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"metry.uk/provext4/MetroTrackerBackend/internal/users/domain"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Create(
	ctx context.Context,
	user *domain.User,
) (*domain.User, error) {

	query := `
		INSERT INTO usuario (
			nombre,
			apepat,
			apemat,
			foto,
			correo,
			password,
			fecharegistro,
			fehcaact,
			idrol,
			idgenero,
			esactivo
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			NOW(),
			NOW(),
			$7,
			$8,
			TRUE
		)
		RETURNING
			id,
			fecharegistro,
			fehcaact,
			esactivo
	`

	err := r.db.QueryRow(
		ctx,
		query,
		user.Nombre,
		user.Apepat,
		user.Apemat,
		user.Foto,
		user.Correo,
		user.Password,
		user.IDRol,
		user.IDGenero,
	).Scan(
		&user.ID,
		&user.FechaRegistro,
		&user.FechaAct,
		&user.EsActivo,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) GetByID(
	ctx context.Context,
	id int32,
) (*domain.User, error) {

	query := `
		SELECT
			u.id,
			u.nombre,
			u.apepat,
			u.apemat,
			u.foto,
			u.correo,
			u.password,
			u.fecharegistro,
			u.fehcaact,
			u.idrol,
			u.idgenero,
			u.esactivo,

			r.id,
			r.nombre,

			g.id,
			g.tipo

		FROM usuario u

		LEFT JOIN roles r
			ON r.id = u.idrol

		LEFT JOIN genero g
			ON g.id = u.idgenero

		WHERE u.id = $1
	`

	user := &domain.User{
		Rol:    &domain.Role{},
		Genero: &domain.Genero{},
	}

	err := r.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&user.ID,
		&user.Nombre,
		&user.Apepat,
		&user.Apemat,
		&user.Foto,
		&user.Correo,
		&user.Password,
		&user.FechaRegistro,
		&user.FechaAct,
		&user.IDRol,
		&user.IDGenero,
		&user.EsActivo,

		&user.Rol.ID,
		&user.Rol.Nombre,

		&user.Genero.ID,
		&user.Genero.Tipo,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}
func (r *UserRepository) GetByEmail(
	ctx context.Context,
	email string,
) (*domain.User, error) {

	query := `
		SELECT
			id,
			nombre,
			apepat,
			apemat,
			foto,
			correo,
			password,
			fecharegistro,
			fehcaact,
			idrol,
			idgenero,
			esactivo
		FROM usuario
		WHERE correo = $1
	`

	user := &domain.User{}

	err := r.db.QueryRow(
		ctx,
		query,
		email,
	).Scan(
		&user.ID,
		&user.Nombre,
		&user.Apepat,
		&user.Apemat,
		&user.Foto,
		&user.Correo,
		&user.Password,
		&user.FechaRegistro,
		&user.FechaAct,
		&user.IDRol,
		&user.IDGenero,
		&user.EsActivo,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}
func (r *UserRepository) List(
	ctx context.Context,
) ([]*domain.User, error) {

	query := `
		SELECT
			u.id,
			u.nombre,
			u.apepat,
			u.apemat,
			u.correo,
			u.fecharegistro,
			u.fehcaact,
			u.idrol,
			u.idgenero,
			u.esactivo,

			r.id,
			r.nombre,

			g.id,
			g.tipo

		FROM usuario u

		LEFT JOIN roles r
			ON r.id = u.idrol

		LEFT JOIN genero g
			ON g.id = u.idgenero

		ORDER BY u.id DESC
	`

	rows, err := r.db.Query(ctx, query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := make([]*domain.User, 0)

	for rows.Next() {

		user := &domain.User{
			Rol:    &domain.Role{},
			Genero: &domain.Genero{},
		}

		err := rows.Scan(
			&user.ID,
			&user.Nombre,
			&user.Apepat,
			&user.Apemat,
			&user.Correo,
			&user.FechaRegistro,
			&user.FechaAct,
			&user.IDRol,
			&user.IDGenero,
			&user.EsActivo,

			&user.Rol.ID,
			&user.Rol.Nombre,

			&user.Genero.ID,
			&user.Genero.Tipo,
		)

		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
func (r *UserRepository) Deactivate(
	ctx context.Context,
	id int32,
) error {

	query := `
		UPDATE usuario
		SET
			esactivo = FALSE,
			fehcaact = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(
		ctx,
		query,
		id,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}
