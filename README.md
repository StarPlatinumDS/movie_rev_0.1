start:
    make run

shutdown:
    Ctrl+C -> make down TODO: change it to one line

migrations:
    migrate create -ext sql -dir migrations -seq <name_of_change>

    # Применить все доступные миграции
    migrate -path migrations -database "postgres://user:pass@localhost:5432/dbname?sslmode=disable" up

    # Откатить последнюю миграцию
    migrate -path migrations -database "postgres://user:pass@localhost:5432/dbname?sslmode=disable" down

    # Посмотреть текущую версию
    migrate -path migrations -database "postgres://user:pass@localhost:5432/dbname?sslmode=disable" version