/*
The love package contains a client library for the Yelp Love API. Refer to Yelp
Love's Readme for information on Yelp Love, including the API and how to set up
an instance: https://github.com/Yelp/love/#api
*/
package love

import "encoding/json"
import "errors"
import "io/ioutil"
import "net/http"
import "net/url"
import "strconv"
import "strings"
import "time"

/*
API Overview

	- GET /api/love - return love(s)
	  - sender: username of sender
	  - recipient: username of recipient
	  - limit: maximum love to return (recommended)
	  - returns JSON list of objects:
	    - sender: username
	    - recipient: username
	    - message: love message
	    - timestamp: time sent, in ISO Format (i.e. YYYY-MM-DD 24:59:59)
	- POST /api/love - send love(s)
	  - sender: username of sender
	  - recipient: username(s) of recipient(s) (comma separated)
	  - message: love message
	  - returns text message, status code indicates success or failure
	- GET /api/autocomplete - return autocomplete suggestions
	  - term: term for autocomplete
	  - returns JSON list of objects:
	    - label: "Full Name (username)"
	    - value: "username"
*/

const LoveGetStatusCode = 200
const LoveCreatedStatusCode = 201
const LoveFailedStatusCode = 418
const LoveBadParamsStatusCode = 422

/*
The Client holds necessary state for creating requests to the Yelp Love API.
ApiKey is generated from the Admin section of the website. BaseUrl should
include the "api" part, but no trailing slash.
EG: https://cwrulove.appspot.com/api
*/
type Client struct {
	ApiKey  string
	BaseUrl string
}

/*
A structure representing a Love.
*/
type Love struct {
	Sender    string
	Recipient string
	Message   string
	Timestamp time.Time
}

/*
Implementing the UnmarshalJSON interface so that we can parse Love.
*/
func (l *Love) UnmarshalJSON(b []byte) error {
	var sender, recipient, message, timestamp string
	var ok bool
	var dict map[string]string
	if err := json.Unmarshal(b, &dict); err != nil {
		return err
	}

	if sender, ok = dict["sender"]; !ok {
		return errors.New("missing key sender")
	}
	if recipient, ok = dict["recipient"]; !ok {
		return errors.New("missing key recipient")
	}
	if message, ok = dict["message"]; !ok {
		return errors.New("missing key message")
	}
	if timestamp, ok = dict["timestamp"]; !ok {
		return errors.New("missing key timestamp")
	}

	var err error
	l.Timestamp, err = time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return errors.New("invalid timestamp encoding")
	}
	l.Recipient = recipient
	l.Message = message
	l.Sender = sender
	return nil
}

/*
A structure representing a Yelp Love user. This is returned from Autocomplete.
*/
type User struct {
	Display  string
	Username string
}

/*
Implements the JSON Unmarshalling interface so that we can load Users from
Autocomplete.
*/
func (u *User) UnmarshalJSON(b []byte) error {
	var ok bool
	var dict map[string]string
	if err := json.Unmarshal(b, &dict); err != nil {
		return err
	}

	if u.Display, ok = dict["label"]; !ok {
		return errors.New("missing key label")
	}
	if u.Username, ok = dict["value"]; !ok {
		return errors.New("missing key value")
	}
	return nil
}

/*
Create a Client. See documentation of Client for more details on the
arguments.
*/
func NewClient(ApiKey string, BaseUrl string) *Client {
	return &Client{
		ApiKey:  ApiKey,
		BaseUrl: BaseUrl,
	}
}

/*
This function retrieves one or more love which were sent from a username, to a
username, up to some limit. Either from or to (but not both) may be an empty
string, indicating that any user is allowed. The limit parameter may be set to
some value <= 0, and a limit will not be requested. However, using a limit and
setting it to some sensible default like 20 is highly encouraged, to avoid
overloading the server. A hard maximum of 2000 love is likely.
*/
func (c *Client) GetLove(from string, to string, limit int64) ([]Love, error) {
	var err error
	var resp *http.Response
	var body []byte
	var loves []Love
	if from == "" && to == "" {
		return nil, errors.New("Must specify at least one of `from` and `to`")
	}
	values := make(url.Values)
	values.Set("api_key", c.ApiKey)
	if from != "" {
		values.Set("sender", from)
	}
	if to != "" {
		values.Set("recipient", to)
	}
	if limit > 0 {
		values.Set("limit", strconv.FormatInt(limit, 10))
	}
	finalUrl := c.BaseUrl + "/love?" + values.Encode()
	if resp, err = http.Get(finalUrl); err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}
	defer resp.Body.Close()
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	}
	if err = json.Unmarshal(body, &loves); err != nil {
		return nil, err
	}
	return loves, nil
}

/*
Send love from a user another user. In this form, the recipient should be a
single string. In fact, the recipient may actually be several usernames
separated by commas.
*/
func (c *Client) SendLove(from string, to string, message string) error {
	var err error
	var resp *http.Response
	finalUrl := c.BaseUrl + "/love"
	contentType := "application-x-www-form-urlencoded"
	values := make(url.Values)
	values.Set("api_key", c.ApiKey)
	values.Set("sender", from)
	values.Set("recipient", to)
	values.Set("message", message)
	body := values.Encode()
	bodyReader := strings.NewReader(body)
	if resp, err = http.Post(finalUrl, contentType, bodyReader); err != nil {
		return err
	}
	if resp.StatusCode != LoveCreatedStatusCode {
		return errors.New(resp.Status)
	}
	return nil
}

/*
Send love from a user to one or more users. In this form, the recipients should
be a slice of strings. The slice should contain at least one username
*/
func (c *Client) SendLoves(from string, to []string, message string) error {
	return c.SendLove(from, strings.Join(to, ","), message)
}

/*
Return completions for a given string. The completions could come from the
username, first, or last name of a user.
*/
func (c *Client) Autocomplete(term string) ([]User, error) {
	var err error
	var resp *http.Response
	var body []byte
	var users []User
	values := make(url.Values)
	values.Set("api_key", c.ApiKey)
	values.Set("term", term)
	finalUrl := c.BaseUrl + "/autocomplete?" + values.Encode()
	if resp, err = http.Get(finalUrl); err != nil {
		return nil, err
	}
	if resp.StatusCode != LoveGetStatusCode {
		return nil, errors.New(resp.Status)
	}
	defer resp.Body.Close()
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	}
	if err = json.Unmarshal(body, &users); err != nil {
		return nil, err
	}
	return users, nil
}
