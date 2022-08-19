package api

const (
	DONE                                   = "Done"
	ERROR_MISSED_ID                        = "Missed id"
	ERROR_NOT_IMPLEMENTED                  = "Not Implemented"
	ERROR_BAD_REQUEST                      = "Bad Request"
	ERROR_INTERNAL_SERVER_ERROR            = "Internal Server Error"
	ERROR_ASSERT_RESULT_TYPE        string = "Unable to assert result type"
	ERROR_MESSAGE_PARSING_BODY_JSON string = "Error during parsing of HTTP request body. Please check it format correctness: missed brackets, double quotes, commas, matching of names and data types and etc"
)
