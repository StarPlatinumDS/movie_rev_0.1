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


pprof:
    # 1. Скачайте профиль сразу после старта
    curl -o heap_before.prof http://127.0.0.1:6060/debug/pprof/heap

    # 2. Сгенерируйте нагрузку (100+ запросов)
    for i in {1..100}; do curl http://localhost:8080/ > /dev/null; done

    # 3. Скачайте профиль после нагрузки
    curl -o heap_after.prof http://127.0.0.1:6060/debug/pprof/heap

    # 4. Сравните: base=до, сравниваем с после
    go tool pprof -base heap_before.prof ./myapp heap_after.prof
    (pprof) top20