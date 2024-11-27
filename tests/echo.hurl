## https://hurl.dev/
#
# requirements:
#  * the echo.yaml config 
#  * github.com/ninlil/http-https-echo as backend

GET http://localhost:8081/sv/any/path
HTTP 200
[Asserts]
jsonpath "$.Method" == "GET"
jsonpath "$.Request" == "/any/path?lang=sv"


GET http://localhost:8081/sv2/any/path
HTTP 402


GET http://localhost:8081/12/any/path
HTTP 402


GET http://localhost:8081/sv/any/path?lang=old
HTTP 200
[Asserts]
jsonpath "$.Method" == "GET"
jsonpath "$.Request" == "/any/path?lang=sv"


PUT http://localhost:8081/sv/any/path?other=anything
HTTP 200
[Asserts]
jsonpath "$.Method" == "PUT"
jsonpath "$.Request" == "/any/path?lang=sv&other=anything"


GET http://localhost:8081/sv/any/path?anything=1234
HTTP 200
[Asserts]
jsonpath "$.Method" == "GET"
jsonpath "$.Request" == "/any/path?anything=1234&lang=sv"


DELETE http://localhost:8081/sv/any/path?lang=old
HTTP 402
# this will match default-http 402 since the DELETE method is listed a a valid match

DELETE http://localhost:8081/unknown
HTTP 402
