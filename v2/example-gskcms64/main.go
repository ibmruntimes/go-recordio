package main

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"unsafe"

	"github.com/ibmruntimes/go-recordio/v2/utils"
)

func getFunc(dll *utils.Dll, str string) uintptr {
	fp, e := dll.Sym(str)
	if e != nil {
		pc, fn, line, _ := runtime.Caller(1)
		log.Fatalf("[FATAL] %s [%s:%s:%d]", e, runtime.FuncForPC(pc).Name(), fn, line)
	}
	return fp
}

var errmap map[uintptr]string = map[uintptr]string{
	0x03353001: "CMSERR_NO_MEMORY",
	0x03353002: "CMSERR_EXT_NOT_SUPPORTED",
	0x03353003: "CMSERR_ALG_NOT_SUPPORTED",
	0x03353004: "CMSERR_BAD_SIGNATURE",
	0x03353005: "CMSERR_CRYPTO_FAILED",
	0x03353006: "CMSERR_IO_CANCELED",
	0x03353007: "CMSERR_IO_ERROR",
	0x03353008: "CMSERR_PWD_MISMATCH",
	0x03353009: "CMSERR_FILE_NOT_FOUND",
	0x0335300a: "CMSERR_DB_CORRUPTED",
	0x0335300b: "CMSERR_MSG_NOT_FOUND",
	0x0335300c: "CMSERR_BAD_HANDLE",
	0x0335300d: "CMSERR_RECORD_DELETED",
	0x0335300e: "CMSERR_RECORD_NOT_FOUND",
	0x0335300f: "CMSERR_INCORRECT_DBTYPE",
	0x03353010: "CMSERR_UPDATE_NOT_ALLOWED",
	0x03353011: "CMSERR_MUTEX_ERROR",
	0x03353012: "CMSERR_BACKUP_EXISTS",
	0x03353013: "CMSERR_DB_EXISTS",
	0x03353014: "CMSERR_RECORD_TOO_BIG",
	0x03353015: "CMSERR_PW_EXPIRED",
	0x03353016: "CMSERR_PW_INCORRECT",
	0x03353017: "CMSERR_ACCESS_DENIED",
	0x03353018: "CMSERR_DB_LOCKED",
	0x03353019: "CMSERR_LENGTH_TOO_SMALL",
	0x0335301a: "CMSERR_NO_PRIVATE_KEY",
	0x0335301b: "CMSERR_BAD_LABEL",
	0x0335301c: "CMSERR_LABEL_NOT_UNIQUE",
	0x0335301d: "CMSERR_RECTYPE_NOT_VALID",
	0x0335301e: "CMSERR_DUPLICATE_CERTIFICATE",
	0x0335301f: "CMSERR_BAD_BASE64_ENCODING",
	0x03353020: "CMSERR_BAD_ENCODING",
	0x03353021: "CMSERR_NOT_YET_VALID",
	0x03353022: "CMSERR_EXPIRED",
	0x03353023: "CMSERR_NAME_NOT_SUPPORTED",
	0x03353024: "CMSERR_ISSUER_NOT_FOUND",
	0x03353025: "CMSERR_PATH_TOO_LONG",
	0x03353026: "CMSERR_INCORRECT_KEY_USAGE",
	0x03353027: "CMSERR_ISSUER_NOT_CA",
	0x03353028: "CMSERR_FMT_NOT_SUPPORTED",
	0x03353029: "CMSERR_ALG_NOT_AVAILABLE",
	0x0335302a: "CMSERR_RECTYPE_CHANGED",
	0x0335302b: "CMSERR_SUBJECT_CHANGED",
	0x0335302c: "CMSERR_PUBLIC_KEY_CHANGED",
	0x0335302d: "CMSERR_DEFAULT_KEY_CHANGED",
	0x0335302e: "CMSERR_SIGNED_CERTS",
	0x0335302f: "CMSERR_CERT_CHAIN_NOT_TRUSTED",
	0x03353030: "CMSERR_KEY_MISMATCH",
	0x03353031: "CMSERR_SIGNER_NOT_FOUND",
	0x03353032: "CMSERR_CONTENT_NOT_SUPPORTED",
	0x03353033: "CMSERR_RECIPIENT_NOT_FOUND",
	0x03353034: "CMSERR_BAD_KEY_SIZE",
	0x03353035: "CMSERR_BAD_KEY_PARITY",
	0x03353036: "CMSERR_WEAK_KEY",
	0x03353037: "CMSERR_BAD_IV_SIZE",
	0x03353038: "CMSERR_BAD_BLOCK_SIZE",
	0x03353039: "CMSERR_BAD_BLOCK_FORMAT",
	0x0335303a: "CMSERR_NO_INVERSE",
	0x0335303b: "CMSERR_LDAP",
	0x0335303c: "CMSERR_LDAP_NOT_AVAILABLE",
	0x0335303d: "CMSERR_BAD_DIGEST_SIZE",
	0x0335303e: "CMSERR_BAD_FILENAME",
	0x0335303f: "CMSERR_OPEN_FAILED",
	0x03353040: "CMSERR_SELF_SIGNED_NOT_FOUND",
	0x03353041: "CMSERR_CERTIFICATE_REVOKED",
	0x03353042: "CMSERR_BAD_ISSUER_NAME",
	0x03353043: "CMSERR_BAD_SUBJECT_NAME",
	0x03353044: "CMSERR_NAME_CONSTRAINTS_VIOLATED",
	0x03353045: "CMSERR_NO_CONTENT_DATA",
	0x03353046: "CMSERR_VERSION_NOT_SUPPORTED",
	0x03353047: "CMSERR_SUBJECT_IS_CA",
	0x03353048: "CMSERR_BAD_DH_PARAMS",
	0x03353049: "CMSERR_BAD_DH_VALUE",
	0x0335304a: "CMSERR_BAD_DSA_PARAMS",
	0x0335304b: "CMSERR_HOST_NOT_VALID",
	0x0335304c: "CMSERR_NO_IMPORT_CERTIFICATE",
	0x0335304d: "CMSERR_CONTENTTYPE_NOT_ALLOWED",
	0x0335304e: "CMSERR_MESSAGEDIGEST_NOT_ALLOWED",
	0x0335304f: "CMSERR_ATTRIBUTE_INVALID_ID",
	0x03353050: "CMSERR_ATTRIBUTE_INVALID_ENUMERATION",
	0x03353051: "CMSERR_CA_NOT_SUPPLIED",
	0x03353052: "CMSERR_BAD_VALIDATION_OPTION",
	0x03353053: "CMSERR_REQUEST_NOT_SUPPLIED",
	0x03353054: "CMSERR_PUBLIC_KEY_INFO_NOT_SUPPLIED",
	0x03353055: "CMSERR_MODULUS_BITS_NOT_SUPPLIED",
	0x03353056: "CMSERR_EXPONENT_NOT_SUPPLIED",
	0x03353057: "CMSERR_PRIVATE_KEY_INFO_NOT_SUPPLIED",
	0x03353058: "CMSERR_MODULUS_NOT_SUPPLIED",
	0x03353059: "CMSERR_PUBLIC_EXPONENT_NOT_SUPPLIED",
	0x0335305A: "CMSERR_PRIVATE_EXPONENT_NOT_SUPPLIED",
	0x0335305B: "CMSERR_PRIME1_NOT_SUPPLIED",
	0x0335305C: "CMSERR_PRIME2_NOT_SUPPLIED",
	0x0335305D: "CMSERR_PRIME_EXPONENT1_NOT_SUPPLIED",
	0x0335305E: "CMSERR_PRIME_EXPONENT2_NOT_SUPPLIED",
	0x0335305F: "CMSERR_COEFFICIENT_NOT_SUPPLIED",
	0x03353060: "CMSERR_BAD_CRL",
	0x03353061: "CMSERR_MULTIPLE_LABEL",
	0x03353062: "CMSERR_MULTIPLE_DEFAULT",
	0x03353063: "CMSERR_DUPLICATE_LABEL_CA_CHAIN",
	0x03353064: "CMSERR_DIGEST_KEY_MISMATCH",
	0x03353065: "CMSERR_RNG",
	0x03353066: "CMSERR_BAD_RNG_OUTPUT",
	0x03353067: "CMSERR_KATPW_FAILED",
	0x03353068: "CMSERR_API_NOT_SUPPORTED",
	0x03353069: "CMSERR_DB_NOT_FIPS",
	0x0335306A: "CMSERR_DB_FIPS_MODE_ONLY",
	0x0335306B: "CMSERR_FIPS_MODE_SWITCH",
	0x0335306C: "CMSERR_FIPS_MODE_EXECUTE_FAILED",
	0x0335306D: "CMSERR_NO_ACCEPTABLE_POLICIES",
	0x0335306E: "CMSERR_BAD_ARG_COUNT",
	0x0335306F: "CMSERR_REQUIRED_EXT_MISSING",
	0x03353070: "CMSERR_BAD_EXT_DATA",
	0x03353071: "CMSERR_CRITICAL_EXT_INCORRECT",
	0x03353072: "CMSERR_DUPLICATE_EXTENSION",
	0x03353073: "CMSERR_DISTRIBUTION_POINTS",
	0x03353074: "CMSERR_FIPS_KEY_PAIR_CONSISTENCY",
	0x03353075: "CMSERR_TASK_MODE_REQUIRED",
	0x03353076: "CMSERR_PRIME_NOT_SUPPLIED",
	0x03353077: "CMSERR_SUB_PRIME_NOT_SUPPLIED",
	0x03353078: "CMSERR_BASE_NOT_SUPPLIED",
	0x03353079: "CMSERR_PRIVATE_VALUE_NOT_SUPPLIED",
	0x0335307A: "CMSERR_PUBLIC_VALUE_NOT_SUPPLIED",
	0x0335307B: "CMSERR_PRIVATE_KEY_NOT_SUPPLIED",
	0x0335307C: "CMSERR_PUBLIC_KEY_NOT_SUPPLIED",
	0x0335307D: "CMSERR_STRUCTURE_TOO_SMALL",
	0x0335307E: "CMSERR_ECURVE_NOT_SUPPORTED",
	0x0335307F: "CMSERR_EC_PARAMETERS_NOT_SUPPLIED",
	0x03353080: "CMSERR_SIGNATURE_NOT_SUPPLIED",
	0x03353081: "CMSERR_BAD_EC_PARAMS",
	0x03353082: "CMSERR_ECURVE_NOT_FIPS_APPROVED",
	0x03353083: "CMSERR_ICSF_NOT_AVAILABLE",
	0x03353084: "CMSERR_ICSF_SERVICE_FAILURE",
	0x03353085: "CMSERR_ICSF_NOT_FIPS",
	0x03353086: "CMSERR_INCORRECT_KEY_TYPE",
	0x03353087: "CMSERR_CRL_EXPIRED",
	0x03353088: "CMSERR_CRYPTO_HARDWARE_NOT_AVAILABLE",
	0x03353089: "CMSERR_ICSF_FIPS_DISABLED",
	0x0335308A: "CMSERR_KATPW_ICSF_FAILED",
	0x0335308B: "CMSERR_BAD_VALIDATE_ROOT_ARG",
	0x0335308C: "CMSERR_PKCS11_LABEL_INVALID",
	0x0335308D: "CMSERR_INCORRECT_KEY_ATTRIBUTE",
	0x0335308E: "CMSERR_PKCS11_OBJECT_NOT_FOUND",
	0x0335308F: "CMSERR_ICSF_FIPS_BAD_ALG_OR_KEY_SIZE",
	0x03353090: "CMSERR_KEY_CANNOT_BE_EXTRACTED",
	0x03353091: "CMSERR_UNICODE_FAILURE",
	0x03353092: "CMSERR_BAD_UTF8_CHARACTER",
	0x03353093: "CMSERR_ICSF_CLEAR_KEY_SUPPORT_NOT_AVAILABLE",
	0x03353094: "CMSERR_OCSP_SIG_REQUIRED",
	0x03353095: "CMSERR_BAD_HTTP_RESPONSE",
	0x03353096: "CMSERR_BAD_OCSP_RESPONSE",
	0x03353097: "CMSERR_OCSP_RESPONDER_ERROR",
	0x03353098: "CMSERR_OCSP_EXPIRED",
	0x03353099: "CMSERR_INVALID_NUMERIC_VALUE",
	0x0335309A: "CMSERR_PKCS7_CMSVERSION_NOT_SUPPORTED",
	0x0335309B: "CMSERR_CERTIFICATE_NOT_SUPPLIED",
	0x0335309C: "CMSERR_OCSP_REQ_ERROR",
	0x0335309D: "CMSERR_MAX_RESPONSE_SIZE_EXCEEDED",
	0x0335309E: "CMSERR_HTTP_IO_ERROR",
	0x0335309F: "CMSERR_BAD_SECURITY_LEVEL_ARG",
	0x033530A0: "CMSERR_EXT_KEY_USAGE_COUNT_IS_INVALID",
	0x033530A1: "CMSERR_EXT_KEY_USAGE_NOT_SUPPLIED",
	0x033530A2: "CMSERR_EXT_KEY_USAGE_COMPARE_FAILED",
	0x033530A3: "CMSERR_EXT_KEY_USAGE_TYPE_IS_INVALID",
	0x033530A4: "CMSERR_CERT_HAS_NO_EXT_KEY_USAGE_EXTENSION",
	0x033530A5: "CMSERR_OCSP_NONCE_CHECK_FAILED",
	0x033530A6: "CMSERR_OCSP_RESPONSE_TIMEOUT",
	0x033530A7: "CMSERR_REVINFO_NOT_YET_VALID",
	0x033530A8: "CMSERR_HOSTNAME_NOT_VALID",
	0x033530A9: "CMSERR_INTERNAL_ERROR",
	0x033530AA: "CMSERR_REQUIRED_BC_EXT_MISSING",
	0x033530AB: "CMSERR_NO_SUBJECT_DN_OR_FRIENDLY_NAME",
	0x033530AC: "CMSERR_BAD_P12_FILE_NAME",
	0x033530AD: "CMSERR_REQUIRED_PARAMETER_NOTSET",
	0x033530AE: "CMSERR_MAX_REV_EXT_LOC_VALUES_EXCEEDED",
	0x033530AF: "CMSERR_HTTP_RESPONSE_TIMEOUT",
	0x033530B0: "CMSERR_LDAP_RESPONSE_TIMEOUT",
	0x033530B1: "CMSERR_UNKNOWN_ERROR",
	0x033530B2: "CMSERR_OCSP_TRY_LATER",
	0x033530B3: "CMSERR_BAD_SIG_ALG_PAIR",
	0x033530B4: "CMSERR_OCSP_RESPONSE_SIGALG_NOT_VALID",
	0x033530B5: "CMSERR_FIPS_MODE_LEVEL_SWITCH",
	0x033530B6: "CMSERR_OCSP_REQUEST_SIGALG_NOT_VALID",
	0x033530B7: "CMSERR_OCSP_RESPONSE_NOT_FOUND",
	0x033530B8: "CMSERR_OCSP_RESPONSE_DUPLICATE_FOUND",
	0x033530B9: "CMSERR_3DES_KEY_PARTS_NOT_UNIQUE",
	0x033530BA: "CMSERR_NON_SUITE_B_CERTIFICATE",
	0x033530BB: "CMSERR_CERT_DB_NOT_SUPPORTED",
	0x033530BC: "CMSERR_BAD_TIME_VALUE",
	0x033530BD: "CMSERR_INVALID_OUTPUT_PARAMETER",
	0x033530BE: "CMSERR_FORMAT_NOT_VALID",
	0x033530BF: "CMSERR_UNSUPPORTED_EXPIRATION",
	0x033530C0: "CMSERR_RSASSA_PSS_DIGEST_ALG_NOT_SUPPORTED",
	0x033530C1: "CMSERR_RSASSA_PSS_MASK_GENERATION_ALG_NOT_SUPPORTED",
	0x033530C2: "CMSERR_ICSF_RSA_PRIVATE_KEY_BAD_TYPE",
	0x033530C3: "CMSERR_INVALID_INPUT_PARAMETER",
	0x033530C4: "CMSERR_INCORRECT_SUITEB_KEY_USAGE",
	0x033530C5: "CMSERR_MSG_CONTAINS_DATA",
	0x033530C6: "CMSERR_STASH_FILE_NOT_FOUND",
	0x033530C7: "CMSERR_STASH_FILE_EMPTY",
}

const (
	GSKDB_EXPORT_DER_BASE64 = 2
)

type GskBuffer struct {
	Length uint32
	_      uint32
	Data   uintptr
}

func ebcdicCstringToString(p uintptr) (ret string) {
	len := utils.Clib(0x0a9, p) // strlen
	if len > 0 {
		b := make([]byte, len)
		utils.Clib(0x094, uintptr(unsafe.Pointer(&b[0])), p, len) // memcpy
		utils.EtoA(b[:])
		ret = string(b)
	} else {
		ret = ""
	}
	return
}
func ebcdicBufferToString(p uintptr, len uint32) (ret string) {
	b := make([]byte, len)
	utils.Clib(0x094, uintptr(unsafe.Pointer(&b[0])), p, uintptr(len)) // memcpy
	utils.EtoA(b[:])
	ret = string(b)
	return
}

func foo() {
	var dll utils.Dll
	e := dll.Open("GSKCMS64")
	if e != nil {
		pc, fn, line, _ := runtime.Caller(1)
		log.Fatalf("[FATAL] %s [%s:%s:%d]", e, runtime.FuncForPC(pc).Name(), fn, line)
	}
	defer func() {
		dll.Close()
	}()
	if utils.Trace {
		dll.ResolveAll()
		fmt.Printf("Function\t\t\t\t\t\t\tAddress\n")
		fmt.Printf("--------\t\t\t\t\t\t\t-------\n")
		for k, v := range dll.Symbols {
			tk := (len(k)) / 8
			if tk > 7 {
				tk = 7
			}
			tabs := 8 - tk
			fmt.Printf("%s%s@%x\n", k, strings.Repeat("\t", tabs), v)
		}
	}
	gsk_open_keyring := getFunc(&dll, "gsk_open_keyring")
	gsk_close_database := getFunc(&dll, "gsk_close_database")
	gsk_get_record_labels := getFunc(&dll, "gsk_get_record_labels")
	gsk_export_certificate := getFunc(&dll, "gsk_export_certificate")
	gsk_free_buffer := getFunc(&dll, "gsk_free_buffer")
	gsk_free_strings := getFunc(&dll, "gsk_free_strings")

	var ringname = []byte("*AUTH*/*")
	utils.AtoE(ringname[:])

	var db uintptr
	var cnt int32
	rc := utils.Cfunc(gsk_open_keyring, uintptr(unsafe.Pointer(&ringname[0])), uintptr(unsafe.Pointer(&db)), uintptr(unsafe.Pointer(&cnt)))

	if rc != 0 {
		fmt.Printf("error %+v\n", errmap[rc])
	}
	defer utils.Cfunc(gsk_close_database, uintptr(unsafe.Pointer(&db)))

	var labels uintptr
	rc = utils.Cfunc(gsk_get_record_labels, db, 0, uintptr(unsafe.Pointer(&cnt)), uintptr(unsafe.Pointer(&labels)))
	if rc != 0 {
		fmt.Printf("error %+v\n", errmap[rc])
	}
	defer utils.Cfunc(gsk_free_strings, uintptr(cnt), labels)
	var i int32
	var gbuf GskBuffer
	for i = 0; i < cnt; i++ {
		label := *(*uintptr)(unsafe.Pointer(labels + uintptr(i<<3)))
		fmt.Printf("%s\n", ebcdicCstringToString(label))
		rc = utils.Cfunc(gsk_export_certificate, db, label, GSKDB_EXPORT_DER_BASE64, uintptr(unsafe.Pointer(&gbuf)))
		if rc != 0 {
			pc, fn, line, _ := runtime.Caller(1)
			log.Fatalf("[FATAL] %s [%s:%s:%d]", e, runtime.FuncForPC(pc).Name(), fn, line)
		}
		defer utils.Cfunc(gsk_free_buffer, uintptr(unsafe.Pointer(&gbuf)))
		fmt.Printf("%s\n", ebcdicBufferToString(gbuf.Data, gbuf.Length))
	}

}
func main() {
	foo()
}