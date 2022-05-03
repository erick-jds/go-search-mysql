package main

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"

	"github.com/bxcodec/faker"
	"github.com/gofiber/fiber"
	"github.com/gofiber/fiber/middleware/cors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Documentos struct {
	Id        uint   `json:"id"`
	Titulo    string `json:"titulo"`
	Descricao string `json:"descricao"`
	Imagem    string `json:"imagem"`
	Preco     int    `json:"preco"`
}

func main() {
	db, err := gorm.Open(mysql.Open("root:@/go_search"), &gorm.Config{})

	if err != nil {
		panic("Could not connect to the database")
	}

	db.AutoMigrate(&Documentos{})

	app := fiber.New()

	app.Use(cors.New())

	app.Post("/api/documentos/populate", func(c *fiber.Ctx) error {
		for i := 0; i < 50; i++ {
			db.Create(&Documentos{
				Titulo:    faker.Word(),
				Descricao: faker.Paragraph(),
				Imagem:    fmt.Sprintf("http://lorempixel.com.br/200/200?%s", faker.UUIDDigit()),
				Preco:     rand.Intn(90) + 10,
			})
		}

		return c.JSON(fiber.Map{
			"message": "success",
		})
	})

	app.Get("/api/documentos/frontend", func(c *fiber.Ctx) error {
		var documentos []Documentos

		db.Find(&documentos)

		return c.JSON(documentos)
	})

	app.Get("/api/documentos/backend", func(c *fiber.Ctx) error {
		var documentos []Documentos

		sql := "SELECT * FROM documentos"
		sqlcount := "SELECT id FROM documentos"

		if s := c.Query("s"); s != "" {
			sql = fmt.Sprintf("%s WHERE titulo LIKE '%%%s%%' OR descricao LIKE '%%%s%%'", sql, s, s)
		}

		if sort := c.Query("sort"); sort != "" {
			sql = fmt.Sprintf("%s ORDER BY preco %s", sql, sort)
		}

		page, _ := strconv.Atoi(c.Query("page", "1"))
		perPage := 9
		var total int64

		db.Raw(sqlcount).Count(&total)

		sql = fmt.Sprintf("%s LIMIT %d OFFSET %d", sql, perPage, (page-1)*perPage)

		db.Raw(sql).Scan(&documentos)

		return c.JSON(fiber.Map{
			"data":      documentos,
			"total":     total,
			"page":      page,
			"last_page": math.Ceil(float64(total / int64(perPage))),
		})
	})

	app.Listen(":8000")
}
