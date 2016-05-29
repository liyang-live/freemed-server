package api

import (
	"fmt"
	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"
)

func init() {
	common.ApiMap["patient"] = common.ApiMapping{
		Authenticated: true,
		RouterFunction: func(r *gin.RouterGroup) {
			r.POST("/search", PatientSearch)
			r.GET("/:id/info", PatientInformation)
			r.GET("/:id/attachments", PatientEmrAttachments)
			r.GET("/:id/attachments/:module", PatientEmrAttachments)
		},
	}
}

type patientEmrAttachmentsResult struct {
	Patient     int64            `db:"patient" json:"patient_id"`
	Module      string           `db:"module" json:"module"`
	Oid         int64            `db:"oid" json:"module_id"`
	Annotation  model.NullString `db:"annotation" json:"annotation"`
	Summary     model.NullString `db:"summary" json:"summary"`
	Stamp       time.Time        `db:"stamp" json:"timestamp"`
	DateMdy     string           `db:"date_mdy" json:"date_mdy"`
	ModuleName  string           `db:"type" json:"module_name"`
	ModuleClass string           `db:"module_namespace" json:"module_namespace"`
	Locked      int              `db:"locked" json:"locked"`
	Id          int              `db:"id" json:"internal_id"`
}

func PatientEmrAttachments(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var query string
	var err error
	var o []patientEmrAttachmentsResult

	module := r.Param("module")
	if module == "" {
		query = "SELECT p.patient AS patient, p.module AS module, p.oid AS oid, p.annotation AS annotation, p.summary AS summary, p.stamp AS stamp, DATE_FORMAT(p.stamp, '%m/%d/%Y') AS date_mdy, m.module_name AS type, m.module_class AS module_namespace, p.locked AS locked, p.id AS id FROM patient_emr p LEFT OUTER JOIN modules m ON m.module_table = p.module WHERE p.patient = ? AND m.module_hidden = 0"
		_, err = model.DbMap.Select(&o, query, id)
	} else {
		query = "SELECT p.patient AS patient, p.module AS module, p.oid AS oid, p.annotation AS annotation, p.summary AS summary, p.stamp AS stamp, DATE_FORMAT(p.stamp, '%m/%d/%Y') AS date_mdy, m.module_name AS type, m.module_class AS module_namespace, p.locked AS locked, p.id AS id FROM patient_emr p LEFT OUTER JOIN modules m ON m.module_table = p.module WHERE p.patient = ? AND p.module = ? AND m.module_hidden = 0"
		_, err = model.DbMap.Select(&o, query, id, module)
	}

	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, o)
	return
}

type patientInformationResult struct {
	Name           string `db:"patient_name" json:"patient_name"`
	Id             string `db:"patient_id" json:"patient_id"`
	DateOfBirth    string `db:"date_of_birth" json:"date_of_birth"`
	Language       string `db:'language" json:"language"`
	DateOfBirthMDY string `db:"date_of_birth_mdy" json:"date_of_birth_mdy"`
	Age            string `db:"age" json:"age"`
	Address1       string `db:"address_line_1" json:"address_line_1"`
	Address2       string `db:"address_line_2" json:"address_line_2"`
	HasAllergy     bool   `db:"hasallergy" json:"hasallergy"`
	City           string `db:"city" json:"city"`
	State          string `db:"state" json:"state"`
	Postal         string `db:"postal" json:"postal"`
	Csz            string `db:"csz" json:"csz"`
	//model.PatientModel
	Pcp      model.NullString `db:"pcp" json:"pcp"`
	Facility model.NullString `db:"facility" json:"facility"`
	Pharmacy model.NullString `db:"pharmacy" json:"pharmacy"`
}

func PatientInformation(r *gin.Context) {
	id := r.Param("id")
	if id == "" {
		r.AbortWithStatus(http.StatusBadRequest)
		return
	}
	query := "SELECT " +
		"CONCAT( p.ptlname, ', ', p.ptfname, IF(NOT ISNULL(p.ptmname), CONCAT(' ', p.ptmname), '') ) AS patient_name" +
		", p.ptid AS patient_id" +
		", p.ptdob AS date_of_birth" +
		", p.ptprimarylanguage AS language" +
		", DATE_FORMAT(p.ptdob, '%m/%d/%Y') AS date_of_birth_mdy" +
		", CASE WHEN ( ( TO_DAYS(NOW()) - TO_DAYS(p.ptdob) ) / 365) >= 2 THEN CONCAT(FLOOR( ( TO_DAYS(NOW()) - TO_DAYS(p.ptdob) ) / 365),' years') ELSE CONCAT(FLOOR( ( TO_DAYS(NOW()) - TO_DAYS(p.ptdob) ) / 30),' months') END AS age" +
		", pa.line1 AS address_line_1" +
		", pa.line2 AS address_line_2" +
		", pa.city AS city" +
		", pa.stpr AS state" +
		", pa.postal AS postal" +
		", CONCAT( pa.city, ', ', pa.stpr, ' ', pa.postal ) AS csz" +
		", CASE WHEN p.id IN ( SELECT al.patient FROM allergies al WHERE al.patient=? AND active = 'active' ) THEN 'true' ELSE 'false' END AS hasallergy" +
		//", p.* " +
		", CONCAT( phy.phylname, ', ', phy.phyfname, ' ', phy.phymname ) AS pcp" +
		", CONCAT( fac.psrname, ' (', fac.psrcity, ', ', fac.psrstate,')' ) AS facility" +
		", CONCAT( ph.phname, ' (', ph.phcity, ', ', ph.phstate,')' ) AS pharmacy " +
		"FROM patient p " +
		"LEFT OUTER JOIN patient_address pa ON ( pa.patient = p.id AND pa.active = TRUE ) " +
		"LEFT OUTER JOIN physician phy ON ( phy.id = p.ptpcp) " +
		"LEFT OUTER JOIN facility fac ON ( fac.id = p.ptprimaryfacility) " +
		"LEFT OUTER JOIN pharmacy ph ON ( ph.id = p.ptpharmacy) " +
		"WHERE p.id = ? GROUP BY p.id"

	var o patientInformationResult
	err := model.DbMap.SelectOne(&o, query, id, id)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, o)
	return
}

type patientSearchResult struct {
	LastName    string `db:"last_name" json:"last_name"`
	FirstName   string `db:"first_name" json:"first_name"`
	MiddleName  string `db:"middle_name" json:"middle_name"`
	PatientId   string `db:"patient_id" json:"patient_id"`
	Age         int64  `db:"age" json:"age"`
	DateOfBirth string `db:"date_of_birth" json:"date_of_birth"`
	Id          int64  `db:"id" json:"id"`
}

func PatientSearch(r *gin.Context) {
	var params gin.H
	if err := r.BindJSON(&params); err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusBadRequest, err)
		return
	}
	log.Printf("PatientSearch(): raw params = %v", params)

	if len(params) < 1 {
		log.Print("PatientSearch(): no usable search parameters found")
		r.AbortWithStatus(http.StatusBadRequest)
	}

	limit := 20

	// Break passed parameters into something usable
	k := make([]string, 0)
	v := make([]interface{}, 0)
	archive := " AND p.ptarchive = 0 "
	for paramName, paramValue := range params {
		log.Printf("PatientSearch(): paramName = %s, paramValue = %v [%s]", paramName, paramValue, reflect.TypeOf(paramValue))
		switch paramName {
		case "age":
			if iv, found := paramValue.(float64); found && iv != 0 {
				k = append(k, fmt.Sprintf("FLOOR( ( TO_DAYS(NOW()) - TO_DAYS(p.ptdob) ) / 365 ) = %d", int64(paramValue.(float64))))
				// no value appended
			}
		case "archive":
			if bv, found := paramValue.(bool); found && bv {
				// If archived patients are included...
				archive = ""
			}
		case "city":
			if sv, found := paramValue.(string); found && sv != "" {
				k = append(k, "pa.city LIKE '%' + ? + '%'")
				v = append(v, paramValue)
			}
		case "dmv":
			if sv, found := paramValue.(string); found && sv != "" {
				k = append(k, "p.dmv LIKE '%' + ? + '%'")
				v = append(v, paramValue)
			}
		case "email":
			if sv, found := paramValue.(string); found && sv != "" {
				k = append(k, "p.pemail LIKE '%' + ? + '%'")
				v = append(v, paramValue)
			}
		case "first_name":
			if sv, found := paramValue.(string); found && sv != "" {
				k = append(k, "p.ptfname LIKE '%' + ? + '%'")
				v = append(v, paramValue)
			}
		case "last_name":
			if sv, found := paramValue.(string); found && sv != "" {
				k = append(k, "p.ptlname LIKE '%' + ? + '%'")
				v = append(v, paramValue)
			}
		case "patient_id":
			if sv, found := paramValue.(string); found && sv != "" {
				k = append(k, "p.ptid LIKE '%' + ? + '%'")
				v = append(v, paramValue)
			}
		case "ssn":
			if sv, found := paramValue.(string); found && sv != "" {
				k = append(k, "p.ssn LIKE '%' + ? + '%'")
				v = append(v, paramValue)
			}
		case "zip":
			if sv, found := paramValue.(string); found && sv != "" {
				k = append(k, "pa.zip LIKE '%' + ? + '%'")
				v = append(v, paramValue)
			}
		default:
			break
		}
	}

	// Build query
	query := fmt.Sprintf("SELECT p.ptlname AS last_name, p.ptfname AS first_name, p.ptmname AS middle_name, p.ptid AS patient_id, FLOOR( ( TO_DAYS(NOW()) - TO_DAYS(p.ptdob) ) / 365 ) AS age, p.ptdob AS date_of_birth, p.id AS id FROM "+model.TABLE_PATIENT+" p LEFT OUTER JOIN "+model.TABLE_PATIENT_ADDRESS+" pa ON p.id = pa.patient WHERE "+strings.Join(k, " AND ")+" AND pa.active = 1 "+archive+" ORDER BY p.ptlname, p.ptfname, p.ptmname LIMIT %d", limit)

	var o []patientSearchResult
	_, err := model.DbMap.Select(&o, query, v...)
	if err != nil {
		log.Print(err.Error())
		r.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	r.JSON(http.StatusOK, o)
	return
}