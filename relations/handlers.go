package relations

import (
	"encoding/json"
	"fmt"
	"net/http"

	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/service-status-go/gtg"
	uuid "github.com/google/uuid"
	"github.com/gorilla/mux"
)

type HttpHandlers struct {
	cypherDriver       Driver
	cacheControlHeader string
}

type ErrorMessage struct {
	Message string `json:"message"`
}

func NewHttpHandlers(cypherDriver Driver, cacheControlHeader string) HttpHandlers {
	return HttpHandlers{cypherDriver, cacheControlHeader}
}

func (hh *HttpHandlers) HealthCheck(neoURL string) fthealth.Check {
	return fthealth.Check{
		BusinessImpact:   "Unable to respond to Relations API requests",
		Name:             "Check connectivity to Neo4j",
		PanicGuide:       "https://runbooks.in.ft.com/upp-relations-api",
		Severity:         1,
		TechnicalSummary: fmt.Sprintf(`Cannot connect to Neo4j (%v). Check that Neo4j instance is up and running`, neoURL),
		Checker:          hh.Checker,
	}
}

func (hh *HttpHandlers) Checker() (string, error) {
	err := hh.cypherDriver.checkConnectivity()
	if err != nil {
		return "Error connecting to Neo4j", err
	}

	return "Connectivity to Neo4j is ok", err
}

func (hh *HttpHandlers) GTG() gtg.Status {
	statusCheck := func() gtg.Status {
		return gtgCheck(hh.Checker)
	}

	return gtg.FailFastParallelCheck([]gtg.StatusChecker{statusCheck})()
}

func gtgCheck(handler func() (string, error)) gtg.Status {
	if _, err := handler(); err != nil {
		return gtg.Status{GoodToGo: false, Message: err.Error()}
	}
	return gtg.Status{GoodToGo: true}
}

func (hh *HttpHandlers) GetContentRelations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	contentUUID := vars["uuid"]

	err := validateUuid(contentUUID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg, jsonErr := json.Marshal(ErrorMessage{fmt.Sprintf("The given uuid is not valid, err=%v", err)})
		if jsonErr != nil {
			w.Write([]byte(fmt.Sprintf("Error message couldn't be encoded in json: , err=%s", jsonErr.Error())))
		} else {
			w.Write([]byte(msg))
		}
		return
	}

	rel, found, err := hh.cypherDriver.findContentRelations(contentUUID)

	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		msg, jsonErr := json.Marshal(ErrorMessage{fmt.Sprintf("Error retrieving relations for %s, err=%v", contentUUID, err)})
		if jsonErr != nil {
			w.Write([]byte(fmt.Sprintf("Error message couldn't be encoded in json: , err=%s", jsonErr.Error())))
		} else {
			w.Write([]byte(msg))
		}
		return
	}
	if !found {
		w.WriteHeader(http.StatusNotFound)
		msg, jsonErr := json.Marshal(ErrorMessage{fmt.Sprintf("No relations found for content with uuid %s", contentUUID)})
		if jsonErr != nil {
			w.Write([]byte(fmt.Sprintf("Error message couldn't be encoded in json: , err=%s", jsonErr.Error())))
		} else {
			w.Write([]byte(msg))
		}
		return
	}

	w.Header().Set("Cache-Control", hh.cacheControlHeader)
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(rel); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg, _ := json.Marshal(ErrorMessage{fmt.Sprintf("Error parsing result for content with uuid %s, err=%v", contentUUID, err)})
		w.Write([]byte(msg))
	}
}

func (hh *HttpHandlers) GetContentCollectionRelations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	contentUUID := vars["uuid"]

	err := validateUuid(contentUUID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg, jsonErr := json.Marshal(ErrorMessage{fmt.Sprintf("The given uuid is not valid, err=%v", err)})
		if jsonErr != nil {
			w.Write([]byte(fmt.Sprintf("Error message couldn't be encoded in json: , err=%s", jsonErr.Error())))
		} else {
			w.Write([]byte(msg))
		}
		return
	}

	rel, found, err := hh.cypherDriver.findContentCollectionRelations(contentUUID)

	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		msg, jsonErr := json.Marshal(ErrorMessage{fmt.Sprintf("Error retrieving relations for %s, err=%v", contentUUID, err)})
		if jsonErr != nil {
			w.Write([]byte(fmt.Sprintf("Error message couldn't be encoded in json: , err=%s", jsonErr.Error())))
		} else {
			w.Write([]byte(msg))
		}
		return
	}
	if !found {
		w.WriteHeader(http.StatusNotFound)
		msg, jsonErr := json.Marshal(ErrorMessage{fmt.Sprintf("No relations found for content collection with uuid %s", contentUUID)})
		if jsonErr != nil {
			w.Write([]byte(fmt.Sprintf("Error message couldn't be encoded in json: , err=%s", jsonErr.Error())))
		} else {
			w.Write([]byte(msg))
		}
		return
	}

	w.Header().Set("Cache-Control", hh.cacheControlHeader)
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(rel); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg, _ := json.Marshal(ErrorMessage{fmt.Sprintf("Error parsing result for content collection with uuid %s, err=%v", contentUUID, err)})
		w.Write([]byte(msg))
	}
}

func validateUuid(contentUUID string) error {
	parsedUUID, err := uuid.Parse(contentUUID)
	if err != nil {
		return err
	}
	if contentUUID != parsedUUID.String() {
		return fmt.Errorf("Parsed UUID (%v) is different than the given uuid (%v).", parsedUUID, contentUUID)
	}
	return nil
}
