package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type data struct{}

func (*data) Clear() {
	So(
		os.RemoveAll("./testdata/migrate"),
		ShouldBeNil,
	)
	So(
		os.MkdirAll("./testdata/migrate", 0777),
		ShouldBeNil,
	)
	Printf("\n      Отчитстка базы куда будут попадать миграции ")
	_ = os.Remove("./testdata/result.db")
}

func (*data) CreateMigrate(
	fileName string,
	fileData string,
) {
	Printf("\n       %s ", fileName)
	err := ioutil.WriteFile(
		fileName,
		[]byte(fileData),
		0644,
	)
	So(err, ShouldBeNil)
}

func (*data) CreateLocks(
	locks ...string,
) {
	Printf("\n      Заполняем список уже установленных миграций ")
	data := strings.Join(locks, "\n")
	So(
		ioutil.WriteFile("./testdata/locks.db", []byte(data+"\n"), 0777),
		ShouldBeNil,
	)
}

func (*data) RunMigrator(
	dir string,
	ext string,
) {
	Printf("\n      Запускаем мигратор ")
	m, err := NewMigrator(
		dir,
		"./testdata/db.sh %s ./testdata/result.db",
		"./testdata/locks.db",
		ext,
	)
	So(err, ShouldBeNil)
	err = m.Do()
	So(err, ShouldBeNil)
}

func (*data) CheckLocks(
	locks ...string,
) {
	Printf("\n        Проверяем lock файл ")
	data, err := ioutil.ReadFile("./testdata/locks.db")
	So(err, ShouldBeNil)
	So(data, ShouldNotBeNil)
	So(
		string(data),
		ShouldEqual,
		strings.Join(locks, "\n")+"\n",
	)
}

func (*data) CheckResults(
	results ...string,
) {
	Printf("\n      Проверяем результат миграции ")
	data, err := ioutil.ReadFile("./testdata/result.db")
	So(err, ShouldBeNil)
	So(data, ShouldNotBeNil)
	So(
		string(data),
		ShouldEqual,
		strings.Join(results, "\n")+"\n",
	)
}

func TestMigrate(t *testing.T) {
	Convey("Проверяем работу мигратора", t, func() {
		Convey("Каталог с файлами", func() {
			test := &data{}
			test.Clear()
			Printf("\n      Подготовка файлов миграции ")
			for i := 1; i < 11; i++ {
				test.CreateMigrate(
					fmt.Sprintf("./testdata/migrate/%02d.sql", i),
					fmt.Sprintf("migrate - %02d.sql", i),
				)
			}
			test.CreateLocks(
				"01.sql",
				"02.sql",
				"05.sql",
			)
			test.RunMigrator(
				"./testdata/migrate",
				".sql",
			)
			test.CheckLocks(
				"01.sql",
				"02.sql",
				"05.sql",
				"03.sql",
				"04.sql",
				"06.sql",
				"07.sql",
				"08.sql",
				"09.sql",
				"10.sql",
			)
			test.CheckResults(
				"migrate - 03.sql",
				"migrate - 04.sql",
				"migrate - 06.sql",
				"migrate - 07.sql",
				"migrate - 08.sql",
				"migrate - 09.sql",
				"migrate - 10.sql",
			)
		})
		Convey("Одиночный файл", func() {
			test := &data{}
			test.Clear()
			Printf("\n      Подготовка файлов миграции ")
			for i := 1; i < 11; i++ {
				test.CreateMigrate(
					fmt.Sprintf("./testdata/migrate/%02d.sql", i),
					fmt.Sprintf("migrate - %02d.sql", i),
				)
			}
			test.CreateLocks(
				"01.sql",
				"02.sql",
				"05.sql",
			)
			test.RunMigrator(
				"./testdata/migrate/07.sql",
				".sql",
			)
			test.CheckLocks(
				"01.sql",
				"02.sql",
				"05.sql",
				"07.sql",
			)
			test.CheckResults(
				"migrate - 07.sql",
			)
		})
	})
}
