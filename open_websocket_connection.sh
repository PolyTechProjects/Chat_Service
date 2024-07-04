CURL_RESULT=$(curl -X POST localhost:1001/login -H 'Content-Type: application/json' -d '{"login":"'"$1"'", "password":"'"$2"'"}' -is)
echo $CURL_RESULT
ACCESS_TOKEN=$(echo $CURL_RESULT | sed 's/\r/\n/g' | grep "Set-Cookie" | sed 's/=/:/g' | sed 's/;/:/g' | cut -d: -f3 | head -n 1)
echo $ACCESS_TOKEN
REFRESH_TOKEN=$(echo $CURL_RESULT | sed 's/\r/\n/g' | grep "Set-Cookie" | sed 's/=/:/g' | sed 's/;/:/g' | cut -d: -f3 | tail -n 1)
echo $REFRESH_TOKEN

USER_ID=$3
echo $USER_ID
AUTH_HEADER='Authorization: Bearer '$ACCESS_TOKEN
echo $AUTH_HEADER
COOKIE_HEADER='Cookie: X-Refresh-Token='$REFRESH_TOKEN
echo $COOKIE_HEADER
USER_ID_HEADER='X-User-Id: '$USER_ID
echo $USER_ID_HEADER
rlwrap websocat ws://localhost:2000/websocket/chat -H "$AUTH_HEADER" -H "$USER_ID_HEADER" -H "$COOKIE_HEADER"
