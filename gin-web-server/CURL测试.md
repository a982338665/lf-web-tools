

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


