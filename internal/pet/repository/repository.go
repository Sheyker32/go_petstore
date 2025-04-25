package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"swagger_petstore/middleware"
	"swagger_petstore/petstore"

	"github.com/go-chi/jwtauth/v5"
	"github.com/lib/pq"

	"github.com/jmoiron/sqlx"
)

type PetsRepository interface {
	AddPet(ctx context.Context, pet petstore.Pet) error
	UpdatePet(ctx context.Context, pet petstore.Pet) error
	FindPetsByStatus(ctx context.Context, status petstore.FindPetsByStatusParams) ([]petstore.Pet, error)
	FindPetsByTags(ctx context.Context, status petstore.FindPetsByTagsParams) ([]petstore.Pet, error)
	DeletePet(ctx context.Context, petId int64, params petstore.DeletePetParams) error
	GetPetById(ctx context.Context, petId int64) (petstore.Pet, error)
	UpdatePetWithForm(ctx context.Context, petId int64, params petstore.UpdatePetWithFormParams) error
}
type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) PetsRepository {
	return &Repository{db: db}
}

func (r *Repository) AddPet(ctx context.Context, pet petstore.Pet) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var categoryID *int64
	if pet.Category != nil && pet.Category.Name != nil {
		err := tx.QueryRow(
			"INSERT INTO categories (name) VALUES ($1) RETURNING id",
			*pet.Category.Name,
		).Scan(&categoryID)
		if err != nil {
			return err
		}
	}

	var petID int64
	err = tx.QueryRow(
		"INSERT INTO pets (name, category_id, photoUrls, status) VALUES ($1, $2, $3, $4) RETURNING id",
		pet.Name, categoryID, pq.Array(pet.PhotoUrls), pet.Status,
	).Scan(&petID)
	if err != nil {
		return err
	}

	if pet.Tags != nil {
		for _, tag := range *pet.Tags {
			if tag.Name != nil {
				var tagID int64

				err := tx.QueryRow(
					"SELECT id FROM tags WHERE name = $1", *tag.Name,
				).Scan(&tagID)

				if err != nil && err != sql.ErrNoRows {
					return err
				}

				if err == sql.ErrNoRows {
					err = tx.QueryRow(
						"INSERT INTO tags (name) VALUES ($1) RETURNING id",
						*tag.Name,
					).Scan(&tagID)
					if err != nil {
						return err
					}
				}

				_, err = tx.Exec(
					"INSERT INTO pet_tags (pet_id, tag_id) VALUES ($1, $2)",
					petID, tagID,
				)
				if err != nil {
					return err
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *Repository) UpdatePet(ctx context.Context, pet petstore.Pet) error {
	if pet.Id == nil {
		return fmt.Errorf("pet ID is required for update")
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
        UPDATE pets 
        SET name = $1, status = $2, photoUrls = $3, category_id = $4
        WHERE id = $5`,
		pet.Name, pet.Status, pq.Array(pet.PhotoUrls), nil, *pet.Id,
	)
	if err != nil {
		return fmt.Errorf("failed to update pet: %v", err)
	}

	if pet.Category != nil {
		categoryID := new(int64)

		if *pet.Category.Id != 0 {
			categoryID = pet.Category.Id
		} else if pet.Category.Name != nil {
			err := tx.QueryRow(`
                INSERT INTO categories (name) 
                VALUES ($1) 
                ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
                RETURNING id`,
				*pet.Category.Name,
			).Scan(&categoryID)
			if err != nil {
				return fmt.Errorf("failed to upsert category: %v", err)
			}
		}

		_, err = tx.Exec(`
            UPDATE pets SET category_id = $1 WHERE id = $2`,
			categoryID, *pet.Id,
		)
		if err != nil {
			return fmt.Errorf("failed to update pet category: %v", err)
		}
	} else {

		_, err = tx.Exec(`
            UPDATE pets SET category_id = NULL WHERE id = $1`,
			*pet.Id,
		)
		if err != nil {
			return fmt.Errorf("failed to clear pet category: %v", err)
		}
	}

	_, err = tx.Exec(`DELETE FROM pet_tags WHERE pet_id = $1`, *pet.Id)
	if err != nil {
		return fmt.Errorf("failed to delete old tags: %v", err)
	}

	if pet.Tags != nil {
		for _, tag := range *pet.Tags {
			var tagID int64

			if *tag.Id != 0 {
				tagID = *tag.Id
			} else if tag.Name != nil {
				// Ищем или создаем тег по имени
				err := tx.QueryRow(`
                    INSERT INTO tags (name) 
                    VALUES ($1) 
                    ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
                    RETURNING id`,
					*tag.Name,
				).Scan(&tagID)
				if err != nil {
					return fmt.Errorf("failed to upsert tag: %v", err)
				}
			}

			if tagID != 0 {
				_, err = tx.Exec(`
                    INSERT INTO pet_tags (pet_id, tag_id) VALUES ($1, $2)`,
					*pet.Id, tagID,
				)
				if err != nil {
					return fmt.Errorf("failed to insert pet tag: %v", err)
				}
			}
		}
	}

	return tx.Commit()
}

func (r *Repository) FindPetsByStatus(ctx context.Context, status petstore.FindPetsByStatusParams) ([]petstore.Pet, error) {
	rows, err := r.db.Query(`
		SELECT p.id, p.name, p.photoUrls, p.category_id, p.status
		FROM pets p
		WHERE p.status = $1
	`, status.Status)

	if err != nil {
		return []petstore.Pet{}, err
	}

	var pets []petstore.Pet
	for rows.Next() {
		var pet petstore.Pet
		var categoryID sql.NullInt64
		var status sql.NullString
		var url []string

		err := rows.Scan(
			&pet.Id,
			&pet.Name,
			pq.Array(&url),
			&categoryID,
			&status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pet row: %v", err)
		}
		pet.PhotoUrls = url

		if status.Valid {
			s := petstore.PetStatus(status.String)
			pet.Status = &s
		}

		if categoryID.Valid {
			category := petstore.Category{
				Id:   new(int64),
				Name: new(string),
			}
			*category.Id = categoryID.Int64
			var name sql.NullString
			err := r.db.QueryRow(`
				SELECT name FROM categories WHERE id = $1
			`, categoryID.Int64).Scan(&name)
			if err != nil && err != sql.ErrNoRows {
				return []petstore.Pet{}, err
			}
			if name.Valid {
				category.Name = &name.String
			}
			pet.Category = &category
		}

		pets = append(pets, pet)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %v", err)
	}

	for i := range pets {
		if pets[i].Id == nil {
			continue
		}

		tagRows, err := r.db.Query(`
            SELECT t.id, t.name 
            FROM tags t
            JOIN pet_tags pt ON t.id = pt.tag_id
            WHERE pt.pet_id = $1`,
			*pets[i].Id,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to query pet tags: %v", err)
		}

		var tagsList []petstore.Tag
		for tagRows.Next() {
			var tag petstore.Tag
			var name sql.NullString
			if err := tagRows.Scan(&tag.Id, &name); err != nil {
				tagRows.Close()
				return nil, fmt.Errorf("failed to scan tag: %v", err)
			}
			if name.Valid {
				tag.Name = &name.String
			}
			tagsList = append(tagsList, tag)
		}
		tagRows.Close()

		if len(tagsList) > 0 {
			pets[i].Tags = &tagsList
		}
	}

	return pets, nil
}

func (r *Repository) FindPetsByTags(ctx context.Context, params petstore.FindPetsByTagsParams) ([]petstore.Pet, error) {
	if len(*params.Tags) == 0 {
		return nil, fmt.Errorf("tags list cannot be empty")
	}

	placeholders := make([]string, len(*params.Tags))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := `
        SELECT p.id, p.name, p.photoUrls, p.status, c.id, c.name
        FROM pets p
        LEFT JOIN categories c ON p.category_id = c.id
        WHERE p.id IN (
            SELECT pt.pet_id
            FROM pet_tags pt
            JOIN tags t ON pt.tag_id = t.id
            WHERE t.name IN (` + strings.Join(placeholders, ", ") + `)
            GROUP BY pt.pet_id
            HAVING COUNT(DISTINCT t.name) = ` + strconv.Itoa(len(placeholders)) + `
        )`
	p := make([]interface{}, len(*params.Tags))
	for i, tag := range *params.Tags {
		p[i] = tag
	}
	// p[len(*params.Tags)] = len(*params.Tags)

	rows, err := r.db.Query(query, p...)
	if err != nil {
		return nil, fmt.Errorf("failed to query pets by tags: %v", err)
	}
	defer rows.Close()

	var pets []petstore.Pet
	for rows.Next() {
		var pet petstore.Pet
		var categoryID sql.NullInt64
		var categoryName sql.NullString
		var status sql.NullString
		var url []string

		err := rows.Scan(
			&pet.Id,
			&pet.Name,
			pq.Array(&url),
			&status,
			&categoryID,
			&categoryName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pet row: %v", err)
		}
		pet.PhotoUrls = url

		if status.Valid {
			s := petstore.PetStatus(status.String)
			pet.Status = &s
		}

		if categoryID.Valid {
			pet.Category = &petstore.Category{
				Id: &categoryID.Int64,
			}
			if categoryName.Valid {
				name := categoryName.String
				pet.Category.Name = &name
			}
		}

		pets = append(pets, pet)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %v", err)
	}

	for i := range pets {
		if pets[i].Id == nil {
			continue
		}

		tagRows, err := r.db.Query(`
            SELECT t.id, t.name 
            FROM tags t
            JOIN pet_tags pt ON t.id = pt.tag_id
            WHERE pt.pet_id = $1`,
			*pets[i].Id,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to query pet tags: %v", err)
		}

		var tagsList []petstore.Tag
		for tagRows.Next() {
			var tag petstore.Tag
			var name sql.NullString
			if err := tagRows.Scan(&tag.Id, &name); err != nil {
				tagRows.Close()
				return nil, fmt.Errorf("failed to scan tag: %v", err)
			}
			if name.Valid {
				tag.Name = &name.String
			}
			tagsList = append(tagsList, tag)
		}
		tagRows.Close()

		if len(tagsList) > 0 {
			pets[i].Tags = &tagsList
		}
	}

	return pets, nil
}

func (r *Repository) DeletePet(ctx context.Context, petId int64, params petstore.DeletePetParams) error {
	var exists bool

	if params.ApiKey != nil {
		apiKey := *params.ApiKey
		_, err := jwtauth.VerifyToken(middleware.TokenAuth, apiKey)
		if err != nil {
			return fmt.Errorf("invalid API key")
		}
	}

	err := r.db.QueryRowxContext(ctx, `
        SELECT EXISTS(SELECT 1 FROM pets WHERE id = $1)`,
		petId,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check pet existence: %v", err)
	}
	if !exists {
		return fmt.Errorf("pet with id %d not found", petId)
	}

	_, err = r.db.ExecContext(ctx, `DELETE FROM pets WHERE id = $1`, petId)
	if err != nil {
		return fmt.Errorf("failed to delete pet: %v", err)
	}

	return nil
}

func (r *Repository) GetPetById(ctx context.Context, petId int64) (petstore.Pet, error) {
	var pet petstore.Pet
	var categoryID sql.NullInt64
	var status sql.NullString
	var url []string

	err := r.db.QueryRow(`
		SELECT p.id, p.name, p.photoUrls, p.category_id, p.status
		FROM pets p
		WHERE p.id = $1
	`, petId).Scan(
		&pet.Id, &pet.Name, pq.Array(&url), &categoryID, &status,
	)

	if err != nil {
		return petstore.Pet{}, err
	}

	pet.PhotoUrls = url

	if status.Valid {
		s := petstore.PetStatus(status.String)
		pet.Status = &s
	}

	if categoryID.Valid {
		category := petstore.Category{
			Id:   new(int64),
			Name: new(string),
		}
		category.Id = &categoryID.Int64
		var name sql.NullString
		err := r.db.QueryRow(`
			SELECT name FROM categories WHERE id = $1
		`, categoryID.Int64).Scan(&name)
		if err != nil && err != sql.ErrNoRows {
			return petstore.Pet{}, err
		}
		if name.Valid {
			category.Name = &name.String
		}
		pet.Category = &category
	}

	rows, err := r.db.Query(`
		SELECT t.id, t.name 
		FROM tags t
		JOIN pet_tags pt ON t.id = pt.tag_id
		WHERE pt.pet_id = $1
	`, petId)
	if err != nil {
		return petstore.Pet{}, err
	}
	defer rows.Close()

	var tags []petstore.Tag
	for rows.Next() {
		var tag petstore.Tag
		var name sql.NullString
		if err := rows.Scan(&tag.Id, &name); err != nil {
			return petstore.Pet{}, err
		}
		if name.Valid {
			tag.Name = &name.String
		}
		tags = append(tags, tag)
	}

	if len(tags) > 0 {
		pet.Tags = &tags
	}

	return pet, nil
}

func (r *Repository) UpdatePetWithForm(ctx context.Context, petId int64, params petstore.UpdatePetWithFormParams) error {
	s := petstore.PetStatus(*params.Status)
	_, err := r.db.Exec("UPDATE pets SET name = $1, status = $2  WHERE id=$3",
		*params.Name, s, petId)
	if err != nil {
		return fmt.Errorf("failed to update pet: %w", err)
	}
	return nil
}
