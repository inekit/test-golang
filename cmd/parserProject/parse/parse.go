package parse

import (
	"bytes"
	"encoding/xml"
	"fmt"

	_ "github.com/lib/pq"
)

type Project struct {
	ID          uint   ``
	Name        string ``
	Description string ``
	Address     string ``
	Building    []Building
}

type Building struct {
	ID      uint   `json:"id" db:"id"`
	Name    string `json:"name" db:"name"`
	Section []Section
}

type Section struct {
	ID   uint   `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	Lot  []Lot
}

type Lot struct {
	ID            uint    `json:"id" db:"id"`
	Floor         uint8   `json:"floor" db:"floor"`
	TotalSquare   float32 `json:"area" db:"ts"`
	LivingSquare  float32 `json:"livingArea" db:"ls"`
	KitchenSquare float32 `json:"kitchenArea" db:"ks"`
	Price         float32 `json:"price" db:"price"`
	LotType       string  `json:"lotType" db:"lt"`
	RoomType      string  `json:"roomType" db:"rt"`
}

type ParseStruct struct {
	Pr_ID             uint    `xml:""`
	Pr_Name           string  `xml:"name"`
	Pr_Description    string  `xml:"description"`
	Pr_Address        string  `xml:"location>address"`
	Bld_ID            uint    `xml:"yandex-building-id" db:"b.id"`
	Bld_Name          string  `xml:"building-name" db:"b.name"`
	Sec_ID            uint    `xml:"building-section" db:"s.id"`
	Sec_Name          string  `xml:"building-state" db:"s.name"`
	Lot_ID            uint    `xml:"internal-id,attr"  db:"l.id"`
	Lot_Floor         uint8   `xml:"floor" db:"l.floor"`
	Lot_TotalSquare   float32 `xml:"area>value" db:"l.ts"`
	Lot_LivingSquare  float32 `xml:"living-space>value" db:"l.ls"`
	Lot_KitchenSquare float32 `xml:"kitchen-space>value" db:"l.ks"`
	Lot_Price         float32 `xml:"price>value" db:"l.price"`
	Lot_Type          string  `xml:"type" db:"l.lt"`
	Lot_RoomType      string  `xml:"category" db:"l.rt"`
}

// Создаем маппинги для каждой структуры. Промежуточное хранилище при парсинге, Sections хранит Lot,
// Building не хранит Section, связь реализуется через BS (хранит ключи для обоих)
type Mappings struct {
	Buildings map[int]Building
	Sections  map[int]Section
	BS        map[int][]int
}

func (ps ParseStruct) Lot() Lot {
	return Lot{ID: ps.Lot_ID, Floor: ps.Lot_Floor, TotalSquare: ps.Lot_TotalSquare, LivingSquare: ps.Lot_LivingSquare,
		KitchenSquare: ps.Lot_KitchenSquare, Price: ps.Lot_Price, LotType: ps.Lot_Type, RoomType: ps.Lot_RoomType}
}

func (ps ParseStruct) Section() Section {
	return Section{ID: ps.Sec_ID, Name: ps.Sec_Name, Lot: []Lot{ps.Lot()}}
}

func (ps ParseStruct) Building() Building {
	return Building{ID: ps.Bld_ID, Name: ps.Bld_Name}
}

// Разбивает полученную при парсинге структуру на маппинги
func (mappings *Mappings) setMappings(ps ParseStruct) {

	// Добавляем новую секцию, или лот в случае если секция существует
	if tSection, ok := mappings.Sections[int(ps.Sec_ID)]; ok {
		tSection.Lot = append(tSection.Lot, ps.Lot())
		mappings.Sections[int(ps.Sec_ID)] = tSection
	} else {
		mappings.Sections[int(ps.Sec_ID)] = ps.Section()
		mappings.BS[int(ps.Bld_ID)] = append(mappings.BS[int(ps.Bld_ID)], int(ps.Sec_ID))
	}

	mappings.Buildings[int(ps.Bld_ID)] = ps.Building()
}

// Формирует структуру project из маппингов
func (mappings *Mappings) Project() Project {
	building := mappings.Buildings
	bs := mappings.BS
	section := mappings.Sections

	var p Project
	for i := range building {
		for bskey := range bs[i] {
			bkey := bs[i][bskey]
			tempB := building[i]
			tempB.Section = append(tempB.Section, section[bkey])
			building[i] = tempB

		}
		p.Building = append(p.Building, building[i])
	}
	return p
}

// Преобразует xml в структуру
func ParseToStruct(f []byte) (Project, error) {
	mappings := Mappings{Buildings: make(map[int]Building), Sections: make(map[int]Section), BS: make(map[int][]int)}
	// Колбэк, вызывается функцией parseEngine для сохранения значений каждого сканированного оффера в mapping-хранилище
	cb := func(ps ParseStruct) error {
		mappings.setMappings(ps)
		return nil
	}

	// Перебираем файл по элементам offer, передавая  промежуточные несвязанные экземпляры  структур в cb
	if err := parseEngine(f, cb); err != nil {
		var p Project
		return p, err
	}
	return mappings.Project(), nil
}

// Cохраняет XML в БД
func ParseToDB(f []byte) error {

	// Cохранение значений каждого сканированного оффера в бд
	storageData := func(ps ParseStruct) error {
		return ConnectDB().SetStruct(ps)
	}

	// Перебираем файл по элементам offer, передавая  промежуточные несвязанные экземпляры  структур в storageData
	return parseEngine(f, storageData)
}

// Парсит xml, сканирует тэги offer, декодирует в общую структуру
// и передает значения в колбэк для дальнейшего преобразования
func parseEngine(f []byte, callback func(ParseStruct) error) error {
	r := bytes.NewReader(f)

	decoder := xml.NewDecoder(r)
	// Перебираем файл по элементам
	cnt := 0
	for {
		tok, err := decoder.Token()
		if err != nil || tok == nil {
			fmt.Println(err)
			break
		}
		switch tp := tok.(type) {
		case xml.StartElement:

			// Выбираем только элементы с тегом offer
			if tp.Name.Local == "offer" {
				cnt++
				//декодирование элемента в структуру
				var parseStruct ParseStruct
				decoder.DecodeElement(&parseStruct, &tp)

				// Возвращаем управление вызвавшей функции для записи данных в бд/структуру
				if er := callback(parseStruct); er != nil {
					return er
				}
			}
		}
	}
	return nil
}

func CheckError(err error) {
	if err != nil {
		fmt.Println(err)
		//panic(err)
	}
}
