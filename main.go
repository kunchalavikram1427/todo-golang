package main

import (
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"os"
)

// Task represents a task in the to-do app
type Task struct {
	ID    int
	Title string
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("To-Do App")

	// Load configurations from config.json
	configFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println("Error opening config file:", err)
		return
	}
	defer configFile.Close()

	var config map[string]map[string]string
	err = json.NewDecoder(configFile).Decode(&config)
	if err != nil {
		fmt.Println("Error decoding config file:", err)
		return
	}

	// Database configuration
	dbConfig := config["database"]

	db := pg.Connect(&pg.Options{
		User:     dbConfig["user"],
		Password: dbConfig["password"],
		Database: dbConfig["databaseName"],
		Addr:     dbConfig["address"],
	})

	// Check database connection
	if err := db.Ping(); err != nil {
		fmt.Println("Error connecting to the database:", err)
		return
	}

	// Create the "tasks" table if not exists
	if err := createSchema(db); err != nil {
		fmt.Println("Error creating schema:", err)
		return
	}

	taskEntry := widget.NewEntry()
	taskList := widget.NewList(
		func() int {
			return len(fetchTasks(db))
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Placeholder")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			label := item.(*widget.Label)
			tasks := fetchTasks(db)
			label.SetText(tasks[id].Title)
		},
	)

	addButton := widget.NewButton("Add", func() {
		title := taskEntry.Text
		if title == "" {
			fmt.Println("Task title cannot be empty")
			return
		}

		newTask := Task{Title: title}
		if _, err := db.Model(&newTask).Insert(); err != nil {
			fmt.Println("Error adding task:", err)
			return
		}

		refreshTaskList(taskList, db)
		taskEntry.Clear()
	})

	removeButton := widget.NewButton("Remove", func() {
		selected := taskList.SelectedID()
		if selected >= 0 {
			task := fetchTasks(db)[selected]
			if _, err := db.Model(&task).Delete(); err != nil {
				fmt.Println("Error removing task:", err)
				return
			}
			refreshTaskList(taskList, db)
		}
	})

	content := container.New(
		layout.NewVBoxLayout(),
		taskEntry,
		taskList,
		container.NewHBox(addButton, removeButton),
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

func createSchema(db *pg.DB) error {
	models := []interface{}{
		(*Task)(nil),
	}

	for _, model := range models {
		err := db.Model(model).CreateTable(&orm.CreateTableOptions{
			IfNotExists: true,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func fetchTasks(db *pg.DB) []Task {
	var tasks []Task
	err := db.Model(&tasks).Select()
	if err != nil {
		fmt.Println("Error fetching tasks:", err)
		return nil
	}
	return tasks
}

func refreshTaskList(list *widget.List, db *pg.DB) {
	list.Refresh()
	list.Select(-1)
	list.Select(len(fetchTasks(db)) - 1)
}
