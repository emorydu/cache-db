package driver

import (
	"encoding/json"
	"fmt"
	"testing"
)

type User struct {
	Name    string
	Age     json.Number
	Contact string
	Company string
	Address Address
}

type Address struct {
	City    string
	State   string
	Country string
	PinCode json.Number
}

func TestCacheDB(t *testing.T) {
	dir := "./"

	db, err := New(dir, nil)
	if err != nil {
		t.Fatal(err)
	}

	employees := []User{
		{"Emory", "24", "23323232", "Google", Address{"Shanghai", "00", "China", "234233"}},
		{"Lin", "22", "23323232", "Google", Address{"Hangzhou", "00", "China", "234233"}},
		{"Gao", "20", "23323232", "Microsoft", Address{"Chengdu", "00", "China", "234233"}},
		{"Smith", "24", "23323232", "Alibaba", Address{"Hangzhou", "00", "China", "234233"}},
		{"John", "24", "23323232", "ByteDance", Address{"Shanghai", "00", "China", "234233"}},
		{"Ross", "24", "23323232", "Tesla", Address{"Shanghai", "00", "China", "234233"}},
	}

	for _, v := range employees {
		if err := db.Write("users", v.Name, User{
			Name:    v.Name,
			Age:     v.Age,
			Contact: v.Contact,
			Company: v.Company,
			Address: v.Address,
		}); err != nil {
			t.Fatal(err)
		}
	}

	records, err := db.ReadAll("users")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(records)

	var users []User

	for _, v := range records {
		user := User{}
		if err = json.Unmarshal([]byte(v), &user); err != nil {
			t.Fatal(err)
		}
		users = append(users, user)
	}
	fmt.Println(users)

	//if err := db.Delete("users", "john"); err != nil {
	//	fmt.Println("Error:", err)
	//}

	if err := db.Delete("users", ""); err != nil {
		fmt.Println("Error:", err)
	}
}
