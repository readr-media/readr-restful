package routes

import (
	"net/http"
	"testing"
	//"github.com/readr-media/readr-restful/config"
	//"github.com/readr-media/readr-restful/models"
)

func TestRouteFilter(t *testing.T) {

	asserter := func(resp string, tc genericTestcase, t *testing.T) {
		//We don't check response content if response is not a pure string
		return
	}

	t.Run("GetFiltered", func(t *testing.T) {
		testcases := []genericTestcase{
			genericTestcase{"MalformedTimeArguments", "GET", `/members/filter?updated_at=1234567890`, ``, http.StatusBadRequest, `{"Error":"Malformed Time Argument"}`},
			genericTestcase{"FilterProjectAllParams", "GET", `/project/filter?max_result=15&page=1&sort=-created_at&id=10&slug=aaabbb&title=t1,t2&description=d1,d2&tag=t1,t2&published_at=$gt:1234567890::$lt:1234567899&updated_at=$gt:1234567890::$lt:1234567899`, ``, http.StatusOK, 0},
			genericTestcase{"FilterPostAllParams", "GET", `/posts/filter?max_result=15&page=1&sort=-created_at&id=10&title=t1,t2&content=c1,c2&author=a1,a2&tag=t1,t2&published_at=$gt:1234567890::$lt:1234567899&updated_at=$gt:1234567890::$lt:1234567899`, ``, http.StatusOK, 0},
			//genericTestcase{"FilterPrijectAllParams", "GET", `/asset/filter?max_result=15&page=1&sort=-created_at&id=10&title=t1,t2&tag=t1,t2&created_at=$gt:1234567890::$lt:1234567899&updated_at=$gt:1234567890::$lt:1234567899`, ``, http.StatusOK, 0}, Moved to asset router
			genericTestcase{"FilterMemberAllParams", "GET", `/members/filter?max_result=15&page=1&sort=-created_at&id=10&mail=yahoo,gmail&nickname=alice,bob&created_at=$gt:1234567890::$lt:1234567899&updated_at=$gt:1234567890::$lt:1234567899`, ``, http.StatusOK, 0},
		}
		for _, tc := range testcases {
			genericDoTest(tc, t, asserter)
		}

	})
}
