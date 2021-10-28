package fcmserverkeyrepo

const (
	// ------ FCM Server Key
	sqlCreateNewFCMServerKey = `INSERT INTO fcm_server_keys (id, app_id, server_key, created_at) 
		VALUES ($1, $2, $3, $4) RETURNING *;`

	sqlSelectFCMServerKey = `SELECT * FROM fcm_server_keys WHERE app_id = $1 ORDER BY created_at DESC;`
)
