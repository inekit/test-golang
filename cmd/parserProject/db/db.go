package internals

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "selectel"
	password = "selectel"
	dbname   = "selectel"
)

type DB struct {
	*sqlx.DB
}

func ConnectDB() *DB {
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sqlx.Open("postgres", psqlconn)
	CheckError(err)
	return &DB{db}
}

func (db *DB) Struct() (Project, error) {
	// Запрашиваем все связанные записи из бд
	var rows []ParseStruct
	if err := db.Select(&rows, `SELECT b.id as "b.id",b.name as "b.name",s.id as "s.id",s.name as "s.name",
	l.id as "l.id", floor as "l.floor", total_square as "l.ts", living_square as "l.ls", kitchen_square as "l.ks", price as "l.price",lot_type as "l.lt",room_type as "l.rt"
	 FROM project.building b,project.section s,project.lot l where l.section_id=s.ID and s.building_id=b.id`); err != nil {
		var p Project
		return p, err
	}

	// Создаем маппинги для каждой структуры. В них хранятся данные во время парсинга XML.
	// Building не хранит Section, связь реализуется через BS (хранит ключи для обоих)
	mappings := Mappings{Buildings: make(map[int]Building), Sections: make(map[int]Section), BS: make(map[int][]int)}
	// Анализируем результаты

	for i := range rows {

		mappings.setMappings(rows[i])
	}

	defer db.Close()

	// Возвращаем преобразованные в экземпляр project данные
	return mappings.Project(), nil
}

func (db *DB) SetStruct(ps ParseStruct) error {

	pQuery := `INSERT INTO project.project (name) values (NULL) RETURNING id`
	bQuery := `INSERT INTO project.building (id, name,project_id) VALUES ($1, $2,$3) ON CONFLICT (id) DO UPDATE SET   (id, name,project_id) = (EXCLUDED.id, EXCLUDED.name, EXCLUDED.project_id)`
	sQuery := `INSERT INTO project.section (id, name, building_id) VALUES ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET   (id, name, building_id) = (EXCLUDED.id, EXCLUDED.name, EXCLUDED.building_id)`
	lQuery := `INSERT INTO project.lot (id, floor, total_square, living_square, kitchen_square, price,lot_type,room_type,section_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) ON CONFLICT (id) DO UPDATE SET   (id, floor, total_square, living_square, kitchen_square, price,lot_type,room_type,section_id) = (EXCLUDED.id, EXCLUDED.floor, EXCLUDED.total_square, EXCLUDED.living_square, EXCLUDED.kitchen_square, EXCLUDED.price,EXCLUDED.lot_type,EXCLUDED.room_type,EXCLUDED.section_id)`

	tx := db.MustBegin()
	var pID int
	tx.QueryRow(pQuery).Scan(&pID)
	tx.MustExec(bQuery, ps.Bld_ID, ps.Bld_Name, pID)
	tx.MustExec(sQuery, ps.Sec_ID, ps.Sec_Name, ps.Bld_ID)
	tx.MustExec(lQuery, ps.Lot_ID, ps.Lot_Floor, ps.Lot_TotalSquare, ps.Lot_LivingSquare,
		ps.Lot_KitchenSquare, ps.Lot_Price, ps.Lot_Type, ps.Lot_RoomType, ps.Sec_ID)
	defer db.Close()
	return tx.Commit()
}
