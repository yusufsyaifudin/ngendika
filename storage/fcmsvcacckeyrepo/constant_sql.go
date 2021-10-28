package fcmsvcacckeyrepo

const (
	// ------ FCM Service Account Key
	sqlCreateNewFCMServiceAccKey = `INSERT INTO fcm_service_account_keys (
			id, app_id, service_account_key, created_at) 
		VALUES ($1, $2, $3, $4) RETURNING *;`

	sqlSelectFCMServiceAccKey = `
		SELECT * FROM fcm_service_account_keys WHERE app_id = $1 ORDER BY created_at DESC;
`
)
