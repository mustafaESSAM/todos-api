package main

import (
	"errors"
	"log"
	"net/http"
	"todos-iris-api/models"

	"github.com/kataras/iris/v12"
	swaggerFiles "github.com/swaggo/files"

	// Use standard http-swagger for maximum compatibility
	swagger "github.com/swaggo/http-swagger"
	"github.com/swaggo/swag/example/basic/docs"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// @title           Todo API with Iris and GORM
// @version         1.0
// @description     A sample server for a Todo list built with the Iris web framework.
// @contact.name    Support
// @license.name    Apache 2.0
// @host            localhost:8080
// @BasePath        /
// @schemes         http

var DB *gorm.DB

// ErrorResponse is the model for error responses used by Swagger documentation
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid input data"`
}

// --- [ Utility Functions ] ---

func InitDatabase() {
	var err error
	DB, err = gorm.Open(sqlite.Open("todos.db"), &gorm.Config{})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Successfully connected to the database.")

	err = DB.AutoMigrate(&models.Todo{})
	if err != nil {
		log.Fatalf("Failed to migrate Todo model: %v", err)
	}
	log.Println("Todo model migrated successfully.")
}

// --- [ Handlers ] ---

// CreateTodo creates a new Todo item (POST /todos)
// @Summary     Create a new Todo item
// @Description Creates a new Todo item.
// @Accept      json
// @Produce     json
// @Param       todo body models.Todo true "Todo item to create"
// @Success     201 {object} models.Todo
// @Failure     400 {object} ErrorResponse "Invalid input data"
// @Failure     500 {object} ErrorResponse "Internal server error"
// @Router      /todos [post]
func CreateTodo(ctx iris.Context) {
	var newTodo models.Todo
	if err := ctx.ReadJSON(&newTodo); err != nil {
		ctx.StopWithJSON(http.StatusBadRequest, iris.Map{"error": "Invalid input data: " + err.Error()})
		return
	}

	result := DB.Create(&newTodo)
	if result.Error != nil {
		ctx.StopWithJSON(http.StatusInternalServerError, iris.Map{"error": "Failed to create Todo item in database."})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(newTodo)
}

// GetTodos retrieves all Todo items (GET /todos)
// @Summary     Get all Todo items
// @Description Retrieves a list of all Todo items.
// @Produce     json
// @Success     200 {array} models.Todo
// @Failure     500 {object} ErrorResponse "Failed to retrieve Todos"
// @Router      /todos [get]
func GetTodos(ctx iris.Context) {
	var todos []models.Todo
	result := DB.Find(&todos)

	if result.Error != nil {
		ctx.StopWithJSON(http.StatusInternalServerError, iris.Map{"error": "Failed to retrieve Todos from database."})
		return
	}
	ctx.JSON(todos)
}

// GetTodoByID retrieves a single Todo item by ID (GET /todos/{id})
// @Summary     Get a single Todo item
// @Description Retrieves a Todo item based on its ID.
// @Produce     json
// @Param       id path int true "Todo ID"
// @Success     200 {object} models.Todo
// @Failure     404 {object} ErrorResponse "Todo item not found"
// @Failure     500 {object} ErrorResponse "Internal server error"
// @Router      /todos/{id} [get]
func GetTodoByID(ctx iris.Context) {
	id := ctx.Params().Get("id")
	var todo models.Todo

	result := DB.First(&todo, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			ctx.StopWithJSON(http.StatusNotFound, iris.Map{"error": "Todo item not found."})
			return
		}
		ctx.StopWithJSON(http.StatusInternalServerError, iris.Map{"error": "Failed to retrieve Todo item from database."})
		return
	}
	ctx.JSON(todo)
}

// UpdateTodo updates an existing Todo item (PUT /todos/{id})
// @Summary     Update an existing Todo item
// @Description Updates the details of an existing Todo item.
// @Accept      json
// @Produce     json
// @Param       id path int true "Todo ID"
// @Param       todo body models.Todo true "Updated Todo object"
// @Success     200 {object} models.Todo
// @Failure     400 {object} ErrorResponse "Invalid input data"
// @Failure     404 {object} ErrorResponse "Todo item not found"
// @Failure     500 {object} ErrorResponse "Internal server error"
// @Router      /todos/{id} [put]
func UpdateTodo(ctx iris.Context) {
	id := ctx.Params().Get("id")
	var existingTodo models.Todo
	result := DB.First(&existingTodo, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			ctx.StopWithJSON(http.StatusNotFound, iris.Map{"error": "Todo item not found."})
			return
		}
		ctx.StopWithJSON(http.StatusInternalServerError, iris.Map{"error": "Database error while fetching item."})
		return
	}

	var updatedTodo models.Todo
	if err := ctx.ReadJSON(&updatedTodo); err != nil {
		ctx.StopWithJSON(http.StatusBadRequest, iris.Map{"error": "Invalid input data."})
		return
	}

	DB.Model(&existingTodo).Updates(updatedTodo)
	DB.First(&existingTodo, id)
	ctx.JSON(existingTodo)
}

// DeleteTodo deletes a specific Todo item (DELETE /todos/{id})
// @Summary     Delete a Todo item
// @Description Deletes a Todo item based on its ID.
// @Produce     json
// @Param       id path int true "Todo ID"
// @Success     204 "No Content"
// @Failure     404 {object} ErrorResponse "Todo item not found"
// @Failure     500 {object} ErrorResponse "Internal server error"
// @Router      /todos/{id} [delete]
func DeleteTodo(ctx iris.Context) {
	id := ctx.Params().Get("id")
	result := DB.Delete(&models.Todo{}, id)

	if result.Error != nil {
		ctx.StopWithJSON(http.StatusInternalServerError, iris.Map{"error": "Failed to delete Todo item from database."})
		return
	}

	if result.RowsAffected == 0 {
		ctx.StopWithJSON(http.StatusNotFound, iris.Map{"error": "Todo item not found."})
		return
	}
	ctx.StatusCode(http.StatusNoContent)
}

// --- [ Main Function ] ---

func main() {
	// 1. Initialize Database
	InitDatabase()

	// 2. Create new IRIS application
	app := iris.New()

	// 3. Group routes under '/todos'
	todosAPI := app.Party("/todos")
	{
		todosAPI.Post("", CreateTodo)
		todosAPI.Get("", GetTodos)
		todosAPI.Get("/{id}", GetTodoByID)
		todosAPI.Put("/{id}", UpdateTodo)
		todosAPI.Delete("/{id}", DeleteTodo)
	}

	// 4. Setup Swagger UI (using standard http-swagger for max compatibility)
	docs.SwaggerInfo.Title = "Todo API"
	//app.Handle("GET", "/swagger/{any:path}", swagger.WrapHandler)

	swaggerHandler := swagger.WrapHandler(swaggerFiles.Handler)
	app.Get("/swagger/{path:path}", func(ctx iris.Context) {
		swaggerHandler.ServeHTTP(ctx.ResponseWriter(), ctx.Request())
	})

	// 5. Start the web server
	log.Fatal(app.Listen(":8080"))
}
