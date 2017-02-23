package love

import "gopkg.in/jarcoal/httpmock.v1"
import "testing"
import "github.com/stretchr/testify/assert"
import "io/ioutil"
import "net/http"
import "net/url"
import "time"

const testApiKey = "abcdefg"
const testBaseUrl = "https://example.com/api"
const testLoveUrl = testBaseUrl + "/love"
const testAutocompleteUrl = testBaseUrl + "/autocomplete"
const singleGetLoveResponse = `[{
"timestamp": "2000-01-01T01:01:01.552636",
"message": "message",
"sender": "hammy",
"recipient": "darwin"
}]`
const twoGetLoveResponse = `[{
"timestamp": "2000-01-01T01:01:01.552636",
"message": "message",
"sender": "hammy",
"recipient": "darwin"
},{
"timestamp": "2000-02-01T01:01:01",
"message": "message",
"sender": "darwin",
"recipient": "hammy"
}]`

func getTestClient() *Client {
	return NewClient(testApiKey, testBaseUrl)
}

func validateParams(t *testing.T, values url.Values, params map[string]string) {
	for k, v := range params {
		assert.Equal(t, values.Get(k), v)
	}
	assert.Equal(t, len(params), len(values))
}

func newGetValidateResponder(t *testing.T, code int, response string,
	params map[string]string) func(*http.Request) (*http.Response, error) {
	return func(req *http.Request) (*http.Response, error) {
		validateParams(t, req.URL.Query(), params)
		return httpmock.NewStringResponse(code, response), nil
	}
}

func newPostValidateResponder(t *testing.T, code int, response string,
	params map[string]string) func(*http.Request) (*http.Response, error) {
	return func(req *http.Request) (*http.Response, error) {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			t.Error(err)
		}
		values, err := url.ParseQuery(string(body))
		if err != nil {
			t.Error(err)
		}
		validateParams(t, values, params)
		return httpmock.NewStringResponse(code, response), nil
	}
}

func TestNewClient(t *testing.T) {
	client := getTestClient()
	assert.Equal(t, client.ApiKey, testApiKey)
	assert.Equal(t, client.BaseUrl, testBaseUrl)
}

func TestGetLoveOnlySender(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := getTestClient()
	params := map[string]string{
		"sender":  "hammy",
		"api_key": testApiKey,
	}

	httpmock.RegisterResponder(
		"GET", testLoveUrl,
		newGetValidateResponder(t, 200, "[]", params),
	)

	loves, err := client.GetLove("hammy", "", 0)
	assert.Nil(t, err)
	assert.Equal(t, len(loves), 0)
}

func TestGetLoveOnlyRecipient(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := getTestClient()
	params := map[string]string{
		"recipient": "darwin",
		"api_key":   testApiKey,
	}

	httpmock.RegisterResponder(
		"GET", testLoveUrl,
		newGetValidateResponder(t, 200, "[]", params),
	)

	loves, err := client.GetLove("", "darwin", 0)
	assert.Nil(t, err)
	assert.Equal(t, len(loves), 0)
}

func TestGetLoveBoth(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := getTestClient()
	params := map[string]string{
		"sender":    "hammy",
		"recipient": "darwin",
		"api_key":   testApiKey,
	}

	httpmock.RegisterResponder(
		"GET", testLoveUrl,
		newGetValidateResponder(t, 200, "[]", params),
	)

	loves, err := client.GetLove("hammy", "darwin", 0)
	assert.Nil(t, err)
	assert.Equal(t, len(loves), 0)
}

func TestGetLoveLimit(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := getTestClient()
	params := map[string]string{
		"sender":    "hammy",
		"recipient": "darwin",
		"limit":     "20",
		"api_key":   testApiKey,
	}

	httpmock.RegisterResponder(
		"GET", testLoveUrl,
		newGetValidateResponder(t, 200, "[]", params),
	)

	loves, err := client.GetLove("hammy", "darwin", 20)
	assert.Nil(t, err)
	assert.Equal(t, len(loves), 0)
}

func TestGetLoveSingleItem(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := getTestClient()

	httpmock.RegisterResponder(
		"GET", testLoveUrl,
		httpmock.NewStringResponder(200, singleGetLoveResponse),
	)

	loves, err := client.GetLove("hammy", "darwin", 20)
	assert.Nil(t, err)
	assert.Equal(t, len(loves), 1)
	assert.Equal(t, loves[0].Sender, "hammy")
	assert.Equal(t, loves[0].Recipient, "darwin")
	assert.Equal(t, loves[0].Message, "message")
	assert.Equal(t, loves[0].Timestamp.Year(), 2000)
	assert.Equal(t, loves[0].Timestamp.Month(), time.Month(1))
	// etc... this is not about testing time.Parse()
}

func TestGetLoveMultiple(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := getTestClient()

	httpmock.RegisterResponder(
		"GET", testLoveUrl,
		httpmock.NewStringResponder(200, twoGetLoveResponse),
	)

	loves, err := client.GetLove("hammy", "darwin", 20)
	assert.Nil(t, err)
	assert.Equal(t, len(loves), 2)
	assert.Equal(t, loves[0].Sender, "hammy")
	assert.Equal(t, loves[1].Sender, "darwin")
}

func TestGetLoveNon200(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := getTestClient()

	httpmock.RegisterResponder(
		"GET", testLoveUrl,
		httpmock.NewStringResponder(LoveBadParamsStatusCode, "message"),
	)

	loves, err := client.GetLove("hammy", "", 0)
	assert.NotNil(t, err)
	assert.Nil(t, loves)
}

func TestSendLoveSingle(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := getTestClient()
	params := map[string]string{
		"api_key":   testApiKey,
		"sender":    "hammy",
		"recipient": "darwin",
		"message":   "message",
	}

	httpmock.RegisterResponder(
		"POST", testLoveUrl,
		newPostValidateResponder(t, 201, "Love sent to darwin!", params),
	)

	err := client.SendLove("hammy", "darwin", "message")
	assert.Nil(t, err)
}

func TestSendLoveMultiple(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := getTestClient()
	params := map[string]string{
		"api_key":   testApiKey,
		"sender":    "hammy",
		"recipient": "darwin,jeremy",
		"message":   "message",
	}

	httpmock.RegisterResponder(
		"POST", testLoveUrl,
		newPostValidateResponder(t, 201, "Love sent to darwin,jeremy!", params),
	)

	err := client.SendLove("hammy", "darwin,jeremy", "message")
	assert.Nil(t, err)
}

func TestSendLovesSingle(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := getTestClient()
	params := map[string]string{
		"api_key":   testApiKey,
		"sender":    "hammy",
		"recipient": "darwin",
		"message":   "message",
	}

	httpmock.RegisterResponder(
		"POST", testLoveUrl,
		newPostValidateResponder(t, 201, "Love sent to darwin!", params),
	)

	err := client.SendLoves("hammy", []string{"darwin"}, "message")
	assert.Nil(t, err)
}

func TestSendLovesMultiple(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := getTestClient()
	params := map[string]string{
		"api_key":   testApiKey,
		"sender":    "hammy",
		"recipient": "darwin,jeremy",
		"message":   "message",
	}

	httpmock.RegisterResponder(
		"POST", testLoveUrl,
		newPostValidateResponder(t, 201, "Love sent to darwin,jeremy!", params),
	)

	err := client.SendLoves("hammy", []string{"darwin", "jeremy"}, "message")
	assert.Nil(t, err)
}

func TestSendLoveNon201(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := getTestClient()

	httpmock.RegisterResponder(
		"POST", testLoveUrl,
		httpmock.NewStringResponder(418, "i'm a litle teapot"),
	)

	err := client.SendLoves("hammy", []string{"darwin", "jeremy"}, "message")
	assert.NotNil(t, err)
}

func TestAutocompleteEmpty(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := getTestClient()
	params := map[string]string{
		"api_key": testApiKey,
		"term":    "ha",
	}

	httpmock.RegisterResponder(
		"GET", testAutocompleteUrl,
		newGetValidateResponder(t, 200, "[]", params),
	)

	users, err := client.Autocomplete("ha")
	assert.Nil(t, err)
	assert.Equal(t, len(users), 0)
}

func TestAutocompleteOne(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := getTestClient()
	params := map[string]string{
		"api_key": testApiKey,
		"term":    "ha",
	}
	response := `[{"label": "label", "value": "value"}]`

	httpmock.RegisterResponder(
		"GET", testAutocompleteUrl,
		newGetValidateResponder(t, 200, response, params),
	)

	users, err := client.Autocomplete("ha")
	assert.Nil(t, err)
	assert.Equal(t, len(users), 1)
	assert.Equal(t, users[0].Display, "label")
	assert.Equal(t, users[0].Username, "value")
}

func TestAutocompleteNon200(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := getTestClient()
	response := `you suck`

	httpmock.RegisterResponder(
		"GET", testAutocompleteUrl,
		httpmock.NewStringResponder(418, response),
	)

	users, err := client.Autocomplete("ha")
	assert.NotNil(t, err)
	assert.Nil(t, users)
}
