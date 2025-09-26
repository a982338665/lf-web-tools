

curl 'http://zjcmss.mioodo.cn/8080/kinovo-performance/orderChange/getList' \
-H 'Accept: application/json, text/plain, */*' \
-H 'Accept-Language: zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7' \
-H 'Access-Control-Allow-Headers: Content-Type, Content-Length, Authorization, Accept, X-Requested-With , yourHeaderFeild' \
-H 'Access-Control-Allow-Methods: PUT,POST,GET,DELETE,OPTIONS' \
-H 'Access-Control-Allow-Origin: *' \
-H 'Authorization: eyJhbGciOiJIUzUxMiJ9.eyJ1c2VyX2lkIjoic3VwZXJhZG1pbiIsIm5hbWUiOiLotoXnuqfnrqHnkIblkZgiLCJ1c2VyX2tleSI6IjUxZTZlNjZhLTg5OGItNDY2Yi1hNjhjLWI0YTI2ZDRjMDUwNyIsImlkIjoic3VwZXJhZG1pbiIsInVzZXJOYW1lIjoic3VwZXJhZG1pbiIsInVzZXJpZCI6InN1cGVyYWRtaW4iLCJ1c2VybmFtZSI6InN1cGVyYWRtaW4ifQ.JYFihy0ljFt7cxFC5y_udfopjJO0hNQHlXD_N0aoUVRT4owBB1agfBjPob1zMHhMpOw6ucypZxoc8n6PRcaveA' \
-H 'Cache-Control: no-cache' \
-H 'Content-Type: application/json' \
-H 'Origin: http://zjcmss.mioodo.cn' \
-H 'Pragma: no-cache' \
-H 'Proxy-Connection: keep-alive' \
-H 'Referer: http://zjcmss.mioodo.cn/' \
-H 'User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36' \
-H 'X-Powered-By: 3.2.1' \
--data-raw '{"pageNo":1,"pageSize":10,"comQueryTxt":null,"isUse":"1","whereJson":[],"orderByColumn":{}}' \
--insecure


curl -X GET https://httpbin.org/get


curl 'https://c-pre.cnbm.com.cn/8080/kinovo-base-data/baseDataUse/getList' \
  -H 'accept: application/json, text/plain, */*' \
  -H 'accept-language: zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7' \
  -H 'access-control-allow-headers: Content-Type, Content-Length, Authorization, Accept, X-Requested-With , yourHeaderFeild' \
  -H 'access-control-allow-methods: PUT,POST,GET,DELETE,OPTIONS' \
  -H 'access-control-allow-origin: *' \
  -H 'authorization: eyJhbGciOiJIUzUxMiJ9.eyJ1c2VyX2lkIjoic3VwZXJhZG1pbiIsIm5hbWUiOiLotoXnuqfnrqHnkIblkZgiLCJ1c2VyX2tleSI6IjljNmIyNzhkLTM0ZDQtNDQyOS05ODBiLWU3MWQyYTJmNWNhZiIsImlkIjoic3VwZXJhZG1pbiIsInVzZXJOYW1lIjoic3VwZXJhZG1pbiIsInVzZXJpZCI6InN1cGVyYWRtaW4iLCJ1c2VybmFtZSI6InN1cGVyYWRtaW4ifQ.f-dx6eZ0ztksiLU-UMsIkLlwjMif7k8D4BAua80yQdg6leraets1YXvs6nyTE-rLZ4CySKXBhdyfGHUXMhU9yA' \
  -H 'cache-control: no-cache' \
  -H 'content-type: application/json' \
  -H 'origin: https://c-pre.cnbm.com.cn' \
  -H 'pragma: no-cache' \
  -H 'priority: u=1, i' \
  -H 'referer: https://c-pre.cnbm.com.cn/' \
  -H 'sec-ch-ua: "Chromium";v="140", "Not=A?Brand";v="24", "Google Chrome";v="140"' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'sec-ch-ua-platform: "Windows"' \
  -H 'sec-fetch-dest: empty' \
  -H 'sec-fetch-mode: cors' \
  -H 'sec-fetch-site: same-origin' \
  -H 'user-agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36' \
  -H 'x-powered-by: 3.2.1' \
  --data-raw '{"tableName":"t_b_nx_loading_discharge","pageNo":1,"pageSize":999,"isUse":"1","whereJson":[{"key":"address_type","symbol":"=","value":"1"}]}'