package relations

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type test struct {
	name             string
	req              *http.Request
	cypherDriverMock *cypherDriverMock
	statusCode       int
	body             string
}

const knownUUID = "f78c1482-a65c-413e-b753-ca3ce3cb84f0"
const successfulContentResponse = `{"curatedRelatedContent":[{"id":"http://id-f78c1482-a65c-413e-b753-ca3ce3cb84f0", "apiUrl":"http://apiurl-f78c1482-a65c-413e-b753-ca3ce3cb84f0"}],
"contains":[{"id":"http://id-f78c1482-a65c-413e-b753-ca3ce3cb84f0", "apiUrl":"http://apiurl-f78c1482-a65c-413e-b753-ca3ce3cb84f0"}],
"containedIn":[{"id":"http://id-f78c1482-a65c-413e-b753-ca3ce3cb84f0", "apiUrl":"http://apiurl-f78c1482-a65c-413e-b753-ca3ce3cb84f0"}]}`
const successfulContentCollectionResponse = `{"containedIn": "f78c1482-a65c-413e-b753-ca3ce3cb84f0",
"contains":["f78c1482-a65c-413e-b753-ca3ce3cb84f0"]}`

func TestGetContentRelationsHandler(t *testing.T) {
	tests := []test{
		{"Success", newRequest("GET", fmt.Sprintf("/content/%s/relations", knownUUID), nil), &cypherDriverMock{contentUUID: knownUUID}, http.StatusOK, successfulContentResponse},
		{"NotFound", newRequest("GET", fmt.Sprintf("/content/%s/relations", "db90a9db-6cb6-4ba0-8648-c0676087aba2"), nil), &cypherDriverMock{contentUUID: knownUUID}, http.StatusNotFound, message("No relations found for content with uuid db90a9db-6cb6-4ba0-8648-c0676087aba2")},
		{"InvalidUuid", newRequest("GET", fmt.Sprintf("/content/%s/relations", "99999"), nil), &cypherDriverMock{contentUUID: knownUUID}, http.StatusBadRequest, message("The given uuid is not valid, err=invalid UUID length: 5")},
		{"ReadError", newRequest("GET", fmt.Sprintf("/content/%s/relations", knownUUID), nil), &cypherDriverMock{contentUUID: knownUUID, failRead: true}, http.StatusServiceUnavailable, message("Error retrieving relations for f78c1482-a65c-413e-b753-ca3ce3cb84f0, err=TEST failing to READ")},
	}

	for _, test := range tests {
		hh := HttpHandlers{test.cypherDriverMock, ""}
		rec := httptest.NewRecorder()
		r := mux.NewRouter()
		r.HandleFunc("/content/{uuid}/relations", hh.GetContentRelations).Methods("GET")
		r.ServeHTTP(rec, test.req)
		assert.True(t, test.statusCode == rec.Code, fmt.Sprintf("%s: Wrong response code, was %d, should be %d", test.name, rec.Code, test.statusCode))
		assert.JSONEq(t, test.body, rec.Body.String(), fmt.Sprintf("%s: Wrong body", test.name))
	}
}

func TestGetContentCollectionRelationsHandler(t *testing.T) {
	tests := []test{
		{"Success", newRequest("GET", fmt.Sprintf("/contentcollection/%s/relations", knownUUID), nil), &cypherDriverMock{contentUUID: knownUUID}, http.StatusOK, successfulContentCollectionResponse},
		{"NotFound", newRequest("GET", fmt.Sprintf("/contentcollection/%s/relations", "db90a9db-6cb6-4ba0-8648-c0676087aba2"), nil), &cypherDriverMock{contentUUID: knownUUID}, http.StatusNotFound, message("No relations found for content collection with uuid db90a9db-6cb6-4ba0-8648-c0676087aba2")},
		{"InvalidUuid", newRequest("GET", fmt.Sprintf("/contentcollection/%s/relations", "99999"), nil), &cypherDriverMock{contentUUID: knownUUID}, http.StatusBadRequest, message("The given uuid is not valid, err=invalid UUID length: 5")},
		{"ReadError", newRequest("GET", fmt.Sprintf("/contentcollection/%s/relations", knownUUID), nil), &cypherDriverMock{contentUUID: knownUUID, failRead: true}, http.StatusServiceUnavailable, message("Error retrieving relations for f78c1482-a65c-413e-b753-ca3ce3cb84f0, err=TEST failing to READ")},
	}

	for _, test := range tests {
		hh := HttpHandlers{test.cypherDriverMock, ""}
		rec := httptest.NewRecorder()
		r := mux.NewRouter()
		r.HandleFunc("/contentcollection/{uuid}/relations", hh.GetContentCollectionRelations).Methods("GET")
		r.ServeHTTP(rec, test.req)
		assert.True(t, test.statusCode == rec.Code, fmt.Sprintf("%s: Wrong response code, was %d, should be %d", test.name, rec.Code, test.statusCode))
		assert.JSONEq(t, test.body, rec.Body.String(), fmt.Sprintf("%s: Wrong body", test.name))
	}
}

func newRequest(method, url string, body []byte) *http.Request {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	return req
}

func message(errMsg string) string {
	return fmt.Sprintf("{\"message\": \"%s\"}\n", errMsg)
}

type cypherDriverMock struct {
	mock.Mock
	contentUUID string
	failRead    bool
}

func (cdm *cypherDriverMock) findContentRelations(contentUUID string) (relations, bool, error) {
	if cdm.failRead {
		return relations{}, false, errors.New("TEST failing to READ")
	}
	if contentUUID == cdm.contentUUID {
		return relations{
			CuratedRelatedContents: []relatedContent{{ID: "http://id-" + contentUUID, APIURL: "http://apiurl-" + contentUUID}},
			Contains:               []relatedContent{{ID: "http://id-" + contentUUID, APIURL: "http://apiurl-" + contentUUID}},
			ContainedIn:            []relatedContent{{ID: "http://id-" + contentUUID, APIURL: "http://apiurl-" + contentUUID}},
		}, true, nil
	}
	return relations{}, false, nil
}

func (cdm *cypherDriverMock) findContentCollectionRelations(contentUUID string) (ccRelations, bool, error) {
	if cdm.failRead {
		return ccRelations{}, false, errors.New("TEST failing to READ")
	}
	if contentUUID == cdm.contentUUID {
		return ccRelations{
			ContainedIn: contentUUID,
			Contains:    []string{contentUUID},
		}, true, nil
	}
	return ccRelations{}, false, nil
}

func (cdm *cypherDriverMock) checkConnectivity() error {
	return nil
}
