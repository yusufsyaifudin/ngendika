package apprepo

const (
	// ------ App
	sqlCreateApp = `INSERT INTO apps (client_id, name, enabled, created_at) VALUES ($1, $2, $3, $4) RETURNING *;`

	sqlGetAppByClientID = `SELECT * FROM apps WHERE LOWER(client_id) = $1 LIMIT 1;`

	// ------ FCM Service Account Key
	sqlCreateNewFCMServiceAccKey = `INSERT INTO fcm_service_account_keys (
			id, app_client_id, service_account_key, created_at) 
		VALUES ($1, $2, $3, $4) RETURNING *;`

	sqlSelectFCMServiceAccKey = `
		SELECT * FROM fcm_service_account_keys WHERE app_client_id = $1 ORDER BY created_at DESC;
`

	// ------ FCM Server Key
	sqlCreateNewFCMServerKey = `INSERT INTO fcm_server_keys (id, app_client_id, server_key, created_at) 
		VALUES ($1, $2, $3, $4) RETURNING *;`

	sqlSelectFCMServerKey = `SELECT * FROM fcm_server_keys WHERE app_client_id = $1 ORDER BY created_at DESC;`
)
