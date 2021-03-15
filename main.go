package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

var (
	versionInfo = "dev"
	dir         = flag.String(
		"dir",
		"project/db/migrate",
		"директория, в которой будут искаться файлы с миграциями",
	)
	command = flag.String(
		"exec",
		"psql -d ${POSTGRES_DB} -h postgres -U ${POSTRGES_USER}",
		"команда для операций с базой данных",
	)
	lock = flag.String(
		"lock",
		"/tmp/migrations.lock",
		"адрес файла со списком уже установленных миграций (выбирайте для каждого проекта свой файл)",
	)
	ext = flag.String(
		"ext",
		".sql",
		"расширение файлов миграций",
	)
)

func main() {
	flag.Parse()

	migrator, err := NewMigrator(
		*dir,
		*command,
		*lock,
		*ext,
	)
	if err != nil {
		panic(err)
	}
	defer migrator.Close()

	files, err := migrator.Files()
	if err != nil {
		panic(err)
	}
	var ok bool
	for _, fileName := range files {
		ok, err = migrator.Migrate(fileName)
		if err != nil {
			panic(err)
		}
		if ok {
			log.Printf("migrate %s", filepath.Base(fileName))
		} else {
			log.Printf("ignore %s", filepath.Base(fileName))
		}
	}
	log.Println("database is up-to-date")
}

// Migrator это механизм миграции
type Migrator struct {
	// dir это адрес каталога с миграционными файлами
	dir string

	// ext это расширение файлов миграций
	ext string

	// exec это команда которой нужно передать файлы миграции
	exec string

	// lock это адрес файла со списком уже выполненных миграций
	lock string

	// lockFile это дискриптер для записи выполненных миграций
	lockFile *os.File

	// locks это список уже установленных миграций
	locks map[string]struct{}
}

// NewMigrator возвращает новый мигратор
func NewMigrator(dir, exec, lock, ext string) (*Migrator, error) {
	switch {
	case dir == "":
		return nil, errors.New("empty dir")
	case exec == "":
		return nil, errors.New("empty exec")
	case lock == "":
		return nil, errors.New("empty lock")
	case ext == "":
		return nil, errors.New("empty ext")
	}
	m := &Migrator{
		dir:   dir,
		ext:   ext,
		exec:  exec,
		lock:  lock,
		locks: make(map[string]struct{}),
	}
	err := m.readLock()
	if err != nil {
		return nil, err
	}
	return m, nil
}

// readLock формирует список уже накатанных миграций
func (m *Migrator) readLock() (err error) {
	// открываем на запись файл со списом выполненных миграций
	m.lockFile, err = os.OpenFile(m.lock, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadFile(m.lock)
	if err != nil {
		return err
	}
	for _, lock := range strings.Split(string(data), "\n") {
		m.locks[strings.TrimSpace(lock)] = struct{}{}
	}
	return nil
}

// Close закрывает открытые файлы
func (m *Migrator) Close() error {
	if m.lockFile == nil {
		return nil
	}
	return m.lockFile.Close()
}

// files возвращает список не выполненных файлом миграции
func (m *Migrator) Files() ([]string, error) {
	file, err := os.Stat(m.dir)
	if err != nil {
		return nil, err
	}
	if !file.IsDir() {
		return []string{
			m.dir,
		}, nil
	}
	var result []string
	files, err := ioutil.ReadDir(m.dir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		result = append(
			result,
			m.dir+string(os.PathSeparator)+file.Name(),
		)
	}
	sort.Strings(result)
	return result, nil
}

// Migrate выполняет миграцию одного файла
func (m *Migrator) Migrate(path string) (bool, error) {
	file, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if file.IsDir() {
		return false, nil
	}
	_, ok := m.locks[file.Name()]
	if ok {
		return false, nil
	}
	if filepath.Ext(file.Name()) != m.ext {
		return false, nil
	}
	command := fmt.Sprintf(m.exec, path)
	commands := strings.Split(command, " ")
	if len(commands) < 2 {
		return false, errors.New("command is too short")
	}
	cmd := exec.Command(commands[0], commands[1:]...)
	var out bytes.Buffer
	cmd.Stderr = &out
	err = cmd.Run()
	if err != nil {
		return false, fmt.Errorf(
			"cannot execute migration file %s: %s, %s\n",
			path,
			err.Error(),
			out.String(),
		)
	}
	_, err = m.lockFile.WriteString(fmt.Sprintf("%s\n", filepath.Base(path)))
	if err != nil {
		return false, fmt.Errorf(
			"cannot write that migration file %s has been executed: %s\n",
			path,
			err.Error(),
		)
	}
	return true, nil
}

// Do выполняет миграцию всех файлов
func (m *Migrator) Do() error {
	files, err := m.Files()
	if err != nil {
		return err
	}
	for _, fileName := range files {
		_, err = m.Migrate(fileName)
		if err != nil {
			return err
		}
	}
	return nil
}
