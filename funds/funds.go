package funds

import (
	"database/sql"
	"context"
	"errors"
	"intive/config"
	"fmt"
)

type Funds struct {
	UserId  int
	Amount float64
}

type Transaction struct {
	Source int
	Destination int
	Amount float64
}

var db *sql.DB

func InitDB(configName string) error{
	connectionString, err := config.NewDBConfig(configName)
	if err != nil {
		fmt.Println("unable to parse configuration file")
		return err
	}

	db, err = sql.Open("mysql", connectionString)
	if err != nil {
		panic(err.Error())
	}

	return err

}
func ListFunds(userId int) (float64, error) {

	row := db.QueryRow("SELECT funds from funds where userid=?", userId)

	var amount float64

	err := row.Scan(&amount)
	if err != nil {
		return 0, err
	}

	return amount, err

}

func TransferFunds(source int, destination int, amount float64) error {
	_, err := ListFunds(destination)
	if err != nil {
		return errors.New("destination user not found")
	}

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	row := tx.QueryRow("SELECT funds from funds where userid=?", source)
	var available float64
	err = row.Scan(&available)
	if err != nil {
		return errors.New("sender user not found")
	}
	if available < amount {
		return errors.New("Insuficient balance")
	}

	_,err = tx.ExecContext(ctx, "update funds set funds=funds+? where userid=?",amount,destination)	
	if err != nil {
		tx.Rollback()
		return err
	}

	_,err = tx.ExecContext(ctx, "update funds set funds=funds - ? where userid=?",amount,source)
	if err != nil {
		tx.Rollback()
		return err
	}

	_,err = tx.ExecContext(ctx, "insert into transfers(source,destination,amount) values(?,?,?)",source, destination,amount)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func ListTransactions(userid int) ([]Transaction, error) {
	rows, err :=  db.Query("SELECT source,destination,amount from transfers where source=? OR destination=?", userid,userid)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var trans Transaction
		err := rows.Scan(&trans.Source, &trans.Destination, &trans.Amount)
		if err != nil {
            return nil, err
		}
		
		transactions = append(transactions, trans)
	}

	return transactions, nil
}