package service

import (
	"context"
	"database/sql"
	"log"
	Error "split/error"
	db "split/postgres"
	SplitWiseRepo "split/splitwise/db"
	"split/splitwise/dto"

	"github.com/lib/pq"
)

type SplitWiseService struct {
}

func (sw *SplitWiseService) AddUser(request *dto.UserAccount) (*dto.Response, error) {
	log.Println("SplitWiseService AddUserAccount Called......")

	if request.EmailId == "" || request.UserPassword == "" {
		log.Println("email or password  is empty")
		return &dto.Response{Message: "Invalid Request"}, Error.ErrInvalidRequest
	}
	if err := SplitWiseRepo.DB.Insert(request); err != nil {
		log.Println("", err)
		return &dto.Response{Message: "Error in inserting"}, err
	}
	return &dto.Response{Message: "sample"}, nil
}

func (sw *SplitWiseService) CreateGroup(request *dto.Request) (*dto.Response, error) {
	log.Println("SplitWiseService CreateGroup Called......")
	log.Println("CreateGroup request", request)

	for _, v := range request.Users_to_add {
		var user dto.UserAccount
		if err := SplitWiseRepo.DB.Find(&user, "emailid", v); err != nil {
			log.Println("errrrrr", err)
			return nil, &Error.Error{Message: "User is not in splitwise"}
		}
	}

	_, err := db.DB.Exec(
		"INSERT INTO UserGroup (group_name, users_to_add, group_expense,created_by) VALUES ($1, $2, $3,$4)",
		request.Group_name,
		pq.Array(request.Users_to_add),
		request.Group_expense,
		request.Created_by,
	)

	if err != nil {
		log.Println("errrrrrrrrrrrrrrrrrrrrrrrr", err)
		return nil, err
	}

	// if err := SplitWiseRepo.DB.Insert(request); err != nil {
	// 	log.Println("error in CreateGroup", err)
	// 	return &dto.Response{Message: "Error in inserting"}, err
	// }
	return &dto.Response{Message: "sample"}, nil
}

func (sw *SplitWiseService) AddUserToGroup(ctx context.Context, request *dto.AddUserToGroup) (*dto.Response, error) {
	var group dto.UserGroup
	email, _ := ctx.Value("useremail").(string)
	query1 := "SELECT created_by FROM UserGroup WHERE group_name = $1 LIMIT 1"
	row := db.DB.QueryRow(query1, request.GroupName)
	if row == nil {
		return nil, &Error.Error{Message: "no such group"}
	}
	if err := row.Scan(&group.Created_by); err != nil {
		if err == sql.ErrNoRows {
			log.Println("No group found with the specified name:", request.GroupName)
			return &dto.Response{Message: "No group found"}, nil
		}
		log.Println("Error scanning row in UserGroup:", err)
		return &dto.Response{Message: "Error scanning data"}, err
	}
	if group.Created_by != email {
		log.Println("only the user who created the group can only add the user")
		return nil, &Error.Error{Message: "only the user who created can only add the user"}
	}
	var user dto.UserAccount
	// Check if the user exists in Splitwise
	if err := SplitWiseRepo.DB.Find(&user, "emailid", request.EmailId); err != nil {
		log.Println("User not found:", err)
		return nil, &Error.Error{Message: "User is not in Splitwise"}
	}
	// Update the UserGroup record to add the email to the Users array
	query := `
        UPDATE UserGroup
        SET users_to_add = array_append(users_to_add, $1)
        WHERE group_name = $2
    `

	// Execute the update query, passing the email and group ID
	if _, err := db.DB.Exec(query, request.EmailId, request.GroupName); err != nil {
		log.Println("Error in updating UserGroup:", err)
		return &dto.Response{Message: "Error in updating"}, err
	}

	return &dto.Response{Message: "User added to group successfully"}, nil
}

func (sw *SplitWiseService) Split(request *dto.Split) (*dto.Response, error) {
	var group dto.UserGroup

	// Change the query to use a parameterized statement and limit to 1 record
	query := "SELECT group_id, group_name, users_to_add, group_expense FROM UserGroup WHERE group_name = $1 LIMIT 1"
	row := db.DB.QueryRow(query, request.GroupName)

	// Scan the result into the group variable
	if err := row.Scan(&group.Group_id, &group.Group_name, pq.Array(&group.Users_to_add), &group.Group_expense); err != nil {
		if err == sql.ErrNoRows {
			log.Println("No group found with the specified name:", request.GroupName)
			return &dto.Response{Message: "No group found"}, nil
		}
		log.Println("Error scanning row in UserGroup:", err)
		return &dto.Response{Message: "Error scanning data"}, err
	}

	// Log the fetched group
	log.Println("Fetched Group:", group)

	// Calculate split and insert into userexpense table
	if group.Group_expense > 0 && len(group.Users_to_add) > 0 {
		switch request.SplitType {
		case "equal":
			log.Println("exacttttttttttttttttttttttttttttttttttttt")
			amount := group.Group_expense / (len(group.Users_to_add))

			for _, email := range group.Users_to_add {
				_, err := db.DB.Exec(
					"INSERT INTO userexpense (emailid, amount, group_id,group_name,issettled) VALUES ($1, $2, $3,$4,$5)",
					email, amount, group.Group_id, group.Group_name, false,
				)
				if err != nil {
					log.Println("Error inserting into userexpense:", err)
					return &dto.Response{Message: "Error recording expense"}, err
				}
			}
			query := `	UPDATE UserGroup SET group_expense = 0 WHERE group_name = $1`
			if _, err := db.DB.Exec(query, request.GroupName); err != nil {
				log.Println("Error in updating UserGroup:", err)
				return &dto.Response{Message: "Error in updating"}, err
			}

		case "exact":
			for i, v := range request.Splitarr {
				_, err := db.DB.Exec(
					"INSERT INTO userexpense (emailid, amount, group_id,group_name) VALUES ($1, $2, $3,$4)",
					i, v, group.Group_id, group.Group_name,
				)
				if err != nil {
					log.Println("Error inserting into userexpense:", err)
					return &dto.Response{Message: "Error recording expense"}, err
				}
			}
			query := `	UPDATE UserGroup SET group_expense = 0 WHERE group_name = $1`
			if _, err := db.DB.Exec(query, request.GroupName); err != nil {
				log.Println("Error in updating UserGroup:", err)
				return &dto.Response{Message: "Error in updating"}, err
			}
		case "percentage":
			for i, v := range request.Splitarr {
				amount := (float64(v) / 100.0) * float64(group.Group_expense)
				log.Println("amounttttttttttttttttttttttttttt", amount, v, group.Group_expense)
				_, err := db.DB.Exec(
					"INSERT INTO userexpense (emailid, amount, group_id,group_name) VALUES ($1, $2, $3,$4)",
					i, amount, group.Group_id, group.Group_name,
				)
				if err != nil {
					log.Println("Error inserting into userexpense:", err)
					return &dto.Response{Message: "Error recording expense"}, err
				}
			}
			query := `	UPDATE UserGroup SET group_expense = 0 WHERE group_name = $1`
			if _, err := db.DB.Exec(query, request.GroupName); err != nil {
				log.Println("Error in updating UserGroup:", err)
				return &dto.Response{Message: "Error in updating"}, err
			}

		}
	}

	return &dto.Response{Message: "Expenses split successfully"}, nil
}

func (sw *SplitWiseService) Payment(ctx context.Context, request *dto.Pay) (*dto.Response, error) {
	email, _ := ctx.Value("useremail").(string)
	// Update the UserGroup record to add the email to the Users array
	log.Println("emaillllllllllllllllll", email)
	query := `
		UPDATE UserExpense
		SET issettled = true
		WHERE group_name = $1 AND emailid = $2
	`
	// Execute the update query, passing the email and group ID
	if _, err := db.DB.Exec(query, request.GroupName, email); err != nil {
		log.Println("Error in updating UserGroup:", err)
		return &dto.Response{Message: "Error in updating"}, err
	}
	return &dto.Response{Message: "sample"}, nil
}

func (sw *SplitWiseService) DeleteGroup(ctx context.Context, request *dto.UserGroup) (*dto.Response, error) {
	log.Println("DeleteGroup called")
	var group dto.UserGroup
	email, _ := ctx.Value("useremail").(string)
	query1 := "SELECT created_by FROM UserGroup WHERE group_name = $1 LIMIT 1"
	row := db.DB.QueryRow(query1, request.Group_name)
	if row == nil {
		return nil, &Error.Error{Message: "no such group"}
	}
	if err := row.Scan(&group.Created_by); err != nil {
		if err == sql.ErrNoRows {
			log.Println("No group found with the specified name:", request.Group_name)
			return &dto.Response{Message: "No group found"}, nil
		}
		log.Println("Error scanning row in UserGroup:", err)
		return &dto.Response{Message: "Error scanning data"}, err
	}
	log.Println("*******************", group.Created_by, email)
	if group.Created_by != email {
		log.Println("only the user who created the group can only add the user")
		return nil, &Error.Error{Message: "only the user who created can only add the user"}
	}
	// Ensure that a valid GroupId is provided
	if request.Group_name == "" {
		return &dto.Response{Message: "Invalid group "}, nil
	}

	// Assuming group.IsSettled is of type []bool or similar
	var groupIsSettled []bool // or the appropriate type

	query2 := "SELECT issettled FROM UserExpense WHERE group_name = $1"
	rows, err := db.DB.Query(query2, request.Group_name)
	if err != nil {
		log.Println("Error executing query:", err)
		return &dto.Response{Message: "Error executing query"}, err
	}
	defer rows.Close() // Ensure rows are closed after use

	// Iterate through the rows and append results to the slice
	for rows.Next() {
		var isSettled bool // Assuming issettled is of type bool
		if err := rows.Scan(&isSettled); err != nil {
			log.Println("Error scanning row in UserExpense:", err)
			return &dto.Response{Message: "Error scanning data"}, err
		}
		groupIsSettled = append(groupIsSettled, isSettled) // Add to the slice
	}

	// Check for any errors encountered during iteration
	if err = rows.Err(); err != nil {
		log.Println("Error iterating over rows:", err)
		return &dto.Response{Message: "Error iterating over rows"}, err
	}

	// Log or use the results as needed
	log.Println("Settled statuses:", groupIsSettled)
	proceed := true
	for _, v := range groupIsSettled {
		if !v {
			proceed = false
		}
	}
	// Define the DELETE SQL query
	if proceed {
		query := `DELETE FROM UserGroup WHERE group_name = $1`
		log.Println("request.Group_namerequest.Group_namerequest.Group_name", request.Group_name)
		// Execute the DELETE query
		if _, err := db.DB.Exec(query, request.Group_name); err != nil {
			log.Println("Error deleting group:", err)
			return &dto.Response{Message: "Error in deleting group"}, err
		}

		return &dto.Response{Message: "Group deleted successfully"}, nil
	} else {
		return &dto.Response{Message: "Amount not fully paid"}, nil
	}

}
