## dice_roll - gRPC-сервис-игра "Подбрасывание кубика"

#### Установка
1. Скопировать репозиторий, перейти в папку
```
git clone git@github.com:artemmj/dice_roll_v2.git

cd dice_roll_v2/
```
2. Создать файл с переменными окружения .env для докера (есть пример .env_example). При первом запуске создастся новая БД в контейнере, применятся миграции, пробросится volume. В .env_example указаны дефолтные параметры подключения, при необходимости можно поменять.
```
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=postgres
```
3. Выполнить команду ```docker-compose up -d --build``` чтобы собрать контейнеры
4. Чтобы сыграть, используем например Postman, подключимся к localhost:50051 by gRPC и выполним Play:
<img width="949" alt="Снимок экрана 2025-06-01 в 12 58 52" src="https://github.com/user-attachments/assets/0661a32c-6b45-4bd3-9cd3-9dd4be8f3cef" />

По логам видно, что происходит игра с выбором разных генераторов.

<img width="1243" alt="Снимок экрана 2025-06-01 в 13 00 21" src="https://github.com/user-attachments/assets/b30546a4-7974-4df2-b84c-3cbb85e5f6a3" />

5. Можно подключиться к контейнеру БД командой
```docker-compose exec db psql -U postgres```
Далее подключиться к БД postgres
```\c postgres;```
и вывести таблицу с результатами game_results
```select * from game_results;```

<img width="665" alt="Снимок экрана 2025-06-01 в 13 02 38" src="https://github.com/user-attachments/assets/eebcada5-ebec-4842-bd53-3163a143a58c" />

6. Чтобы исполнить тесты, выполните команду при работающем контейнере
```
go test ./tests -count=1 -v
```
