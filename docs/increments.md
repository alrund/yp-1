#### Инкремент 1

Сервер должен быть доступен по адресу: http://localhost:8080.

Сервер должен предоставлять два эндпоинта: POST / и GET /{id}.

Эндпоинт POST / принимает в теле запроса строку URL для сокращения и возвращает ответ с кодом 201
и сокращённым URL в виде текстовой строки в теле.

Эндпоинт GET /{id} принимает в качестве URL-параметра идентификатор сокращённого URL
и возвращает ответ с кодом 307 и оригинальным URL в HTTP-заголовке Location.

Нужно учесть некорректные запросы и возвращать для них ответ с кодом 400.

#### Инкремент 2
Покройте сервис юнит-тестами. Сконцентрируйтесь на покрытии тестами эндпоинтов,
чтобы защитить API сервиса от случайных изменений.

#### Инкремент 3
Вы написали приложение с помощью стандартной библиотеки net/http.
Используя любой пакет (роутер или фреймворк), совместимый с net/http, перепишите ваш код.
Задача направлена на рефакторинг приложения с помощью готовой библиотеки.
Обратите внимание, что необязательно запускать приложение вручную: тесты,
которые вы написали до этого, помогут вам в рефакторинге.

#### Инкремент 4
Добавьте в сервер новый эндпоинт POST /api/shorten,
принимающий в теле запроса JSON-объект {"url":"<some_url>"} и возвращающий в ответ объект {"result":"<shorten_url>"}.

#### Инкремент 5
Добавьте возможность конфигурировать сервис с помощью переменных окружения:
- адрес запуска HTTP-сервера с помощью переменной SERVER_ADDRESS.
- базовый адрес результирующего сокращённого URL с помощью переменной BASE_URL.

#### Инкремент 6
Сохраняйте все сокращённые URL на диск в виде файла. При перезапуске приложения все URL должны быть восстановлены.

Путь до файла должен передаваться в переменной окружения FILE_STORAGE_PATH.

При отсутствии переменной окружения или при её пустом значении вернитесь к хранению сокращённых URL в памяти.

#### Инкремент 7
Поддержите конфигурирование сервиса с помощью флагов командной строки наравне с уже имеющимися переменными окружения:
- флаг -a, отвечающий за адрес запуска HTTP-сервера (переменная SERVER_ADDRESS);
- флаг -b, отвечающий за базовый адрес результирующего сокращённого URL (переменная BASE_URL);
- флаг -f, отвечающий за путь до файла с сокращёнными URL (переменная FILE_STORAGE_PATH).

#### Инкремент 8
Добавьте поддержку gzip в ваш сервис. Научите его:
- принимать запросы в сжатом формате (HTTP-заголовок Content-Encoding);
- отдавать сжатый ответ клиенту, который поддерживает обработку сжатых ответов (HTTP-заголовок Accept-Encoding).

#### Инкремент 9

Добавьте в сервис функциональность аутентификации пользователя.
Сервис должен:
- Выдавать пользователю симметрично подписанную куку, содержащую уникальный идентификатор пользователя, если такой куки не существует или она не проходит проверку подлинности.
- Иметь хендлер GET /api/user/urls, который сможет вернуть пользователю все когда-либо сокращённые им URL в формате:
  ```
  [
     {
       "short_url": "http://...",
       "original_url": "http://..."
     },
     ...
  ]
  ```
- При отсутствии сокращённых пользователем URL хендлер должен отдавать HTTP-статус 204 No Content.
Получить куки запроса можно из поля (*http.Request).Cookie, а установить — методом http.SetCookie.