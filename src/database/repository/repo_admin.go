package repository

import (
	"fmt"
	"time"
	"viral-game-network/src/database"
	dbtype "viral-game-network/src/database/type"
)

func AllAdmin(count int, pager int) ([]dbtype.Admin, int, error) {
	// Get All Admins
	admins, err := database.Query[dbtype.Admin](`
	SELECT * 
	FROM type::table(Admin) 
	ORDER BY date_created DESC 
	LIMIT $ct START $pg;`,
		map[string]interface{}{
			"ct": count,
			"pg": (pager - 1) * count,
		})
	if err != nil {
		return nil, 0, err
	}

	// Count All Admins
	total, err := database.Query[dbtype.Total]("SELECT count() AS total FROM type::table(Admin) GROUP ALL;",
		map[string]interface{}{
		})
	if err != nil {
		return admins, 0, err
	}
	return admins, total[0].Total, nil
}

func GetAdmin(id string) (*dbtype.Admin, error) {
	admins, err := database.Query[dbtype.Admin](
		`SELECT * FROM type::record($id);`,
		map[string]interface{}{
			"id": id,
		},
	)
	if err != nil {
		return nil, err
	}

	if len(admins) == 0 {
		return nil, fmt.Errorf("Admin not found")
	}

	return &admins[0], nil
}

func SelAdmin(info *dbtype.Admin) (*dbtype.Admin, error) {
	// Get admin by name or guid.
	admins, err := database.Query[dbtype.Admin](
		`
		SELECT * 
		FROM type::table(Admin) 
		WHERE 
			(name = $name OR email = $email OR phone = $phone) 
		LIMIT 1;`,
		map[string]interface{}{
			"name": info.Name,
			"email": info.Email,
			"phone": info.Phone,
		},
	)
	if err != nil {
		return nil, err
	}

	if len(admins) == 0 {
		return nil, fmt.Errorf("Admin not found")
	}

	return &admins[0], nil
}

func SetAdmin(id string, admin *dbtype.Admin) (*dbtype.Admin, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	admin.Date_Updated = now
	return database.Update[dbtype.Admin](*admin.ID, admin)
}

func PutAdmin(admin *dbtype.Admin) (*dbtype.Admin, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	admin.Date_Created = now
	return database.Create[dbtype.Admin](admin)
}

func DelAdmin(id string, admin *dbtype.Admin) error {
	return database.Delete(*admin.ID)
}