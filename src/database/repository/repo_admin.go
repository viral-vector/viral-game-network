package repository

import (
	"fmt"
	"time"
	"viral-game-network/src/database"
	dbtype "viral-game-network/src/database/type"
)

func GenAdmin(username string, password string) (*dbtype.Admin, error) {
	admin, err := PutAdmin(&dbtype.Admin{
		Name:     	username,
		Email: 		username,
		Phone: 		"1234567890",
		Password: 	password,
	})
	if err != nil {
		return nil, err
	}
	
	return admin, nil
}

func AllAdmin(count int, pager int, search string) ([]dbtype.Admin, int, error) {
	// Get All Admins
	lQuery := `
	SELECT * 
	FROM type::table(Admin) 
	`
	params := map[string]interface{}{}
	filter := ""

	// Search
	if search != "" {
		filter = " WHERE array::any([name, email], |$value| string::similarity::fuzzy(<string>($value ?? ''), $search) > 0)"
		lQuery += filter
		params["search"] = fmt.Sprintf("%s", search)
	}

	// Ordering
	lQuery = fmt.Sprintf("%s ORDER BY date_created DESC, id ASC", lQuery)

	// Limit and Pagination
	if count > -1 {
		lQuery = fmt.Sprintf("%s LIMIT $ct START $pg", lQuery)
		params["ct"] = count
		params["pg"] = (pager - 1) * count
	}

	// Get Admins.
	admins, err := database.Query[dbtype.Admin](fmt.Sprintf("%s;", lQuery), params)
	if err != nil {
		return nil, 0, err
	}

	// Count all Admins.
	total, err := database.Query[dbtype.Total](
		fmt.Sprintf("SELECT count() AS total FROM type::table(Admin)%s GROUP ALL;", filter),
		params,
	)
	if err != nil || len(total) == 0 {
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