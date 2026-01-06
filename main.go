package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"cloud.google.com/go/firestore"
	"github.com/xuri/excelize/v2"
	"google.golang.org/api/option"
)

func main() {
	mode := flag.String("mode", "", "Pilih mode: backup atau restore")
	flag.Parse()

	if *mode != "backup" && *mode != "restore" {
		fmt.Println("Error: Harap tentukan mode!")
		fmt.Println("Penggunaan: go run main.go -mode=restore ATAU go run main.go -mode=backup")
		return
	}

	ctx := context.Background()
	sa := option.WithAuthCredentialsFile(option.ServiceAccount, "service-account.json")
	db, err := firestore.NewClientWithDatabase(ctx, "interactive-devops-task", "db-client-dev", sa)
	if err != nil {
		fmt.Printf("Gagal koneksi database: %v\n", err)
		return
	}
	defer db.Close()

	if *mode == "restore" {
		doRestore(ctx, db)
	} else {
		doBackup(ctx, db)
	}
}

// Restore Function
func doRestore(ctx context.Context, db *firestore.Client) {
	fmt.Println("Memulai Restore JSON data to Firestore")
	files, _ := filepath.Glob("databasefile/*.json")
	for _, filename := range files {
		raw, _ := os.ReadFile(filename)
		var items []map[string]interface{}
		json.Unmarshal(raw, &items)
		for _, item := range items {
			db.Collection("clients").Doc(fmt.Sprint(item["mid"])).Set(ctx, item)
		}
		fmt.Println("Berhasil restore:", filename)
	}
}

// Backup Function
func doBackup(ctx context.Context, db *firestore.Client) {
	fmt.Println("Memulai Backup")
	docs, _ := db.Collection("clients").Documents(ctx).GetAll()
	var dataList []map[string]interface{}
	xl := excelize.NewFile()

	xl.SetCellValue("Sheet1", "A1", "mid")
	xl.SetCellValue("Sheet1", "B1", "namaClient")

	for i, doc := range docs {
		d := doc.Data()
		dataList = append(dataList, d)
		rowNum := i + 2
		xl.SetCellValue("Sheet1", fmt.Sprintf("A%d", rowNum), d["mid"])
		xl.SetCellValue("Sheet1", fmt.Sprintf("B%d", rowNum), d["namaClient"])
	}

	raw, _ := json.MarshalIndent(dataList, "", "  ")
	os.WriteFile("backup.json", raw, 0644)
	xl.SaveAs("backup.xlsx")
	fmt.Println("Berhasil ekspor backup.json dan backup.xlsx")
}
