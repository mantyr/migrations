# Migrations - утилита для накатывания миграций на базу

### Выполнение миграций

	./main --exec="./testdata/db.sh %s ./testdata/result.db" --dir="./testdata/migrate" --lock="./testdata/locks.db"
	2019/04/01 17:01:21 migrate 10.sql
	2019/04/01 17:01:21 database is up-to-date

	./main --exec="psql -d ${POSTGRES_DB} -h postgres -U ${POSTRGES_USER} -f %s" --dir="./testdata/migrate" --lock="./testdata/locks.db"
	2019/04/01 17:01:21 migrate 10.sql
	2019/04/01 17:01:21 database is up-to-date

Здесь, параметры запуска:
- **dir** - директория, в которой будут искаться файлы с миграциями
- **exec** - произвольная команда для операций с базой: psql, mysql, и другое...
- **lock** - адрес лок файла, в него записываются названия всех установленных миграций

### Пример для .gitlab-ci.yml

	before_script:
	  - go install github.com/mantyr/migrations
	  - PGPASSWORD="$POSTGRES_PASSWORD" migrations --lock="/tmp/lock.db" --exec="psql -d ${POSTGRES_DB} -h postgres -U ${POSTRGES_USER} -f %s" --dir="./sql/schema.sql"
	  - PGPASSWORD="$POSTGRES_PASSWORD" migrations --lock="/tmp/lock.db" --exec="psql -d ${POSTGRES_DB} -h postgres -U ${POSTRGES_USER} -f %s" --dir="./sql/migrate"
	services:
	  - postgres:13

### Принцип работы
Файлы с миграциями скопированы или слинкованы в папку, указанную в параметрах запуска

    .
    ├── sqlchangesets 
    │   ├── 0000_....sql
    │   ├── 0001_....sql
    │   ├── 0002_....sql     

Мигратор читает все файлы по порядку и выполняет на базе с помощью строки подключения, переданной при запуске.

Выполненные миграции записываются в файл адрес которого указывается в **lock**.
Соответственно, файлы, перечисленные ранее в этом файле, не выполняются.

При первой ошибке скрипт останавливается.

*Пароль для подключения к базе* скриптом не указывается и не проверяется. 
Следует указать его через переменные окружения `(PGPASSWORD)`, через файлы конфига или передать в строке подключения

##### Пожелания к оформлению миграций:
Подумать о возможности выполнения миграции несколько раз: добавить `if no exists, create or replace, drop if exists` и другие подобные проверки.
