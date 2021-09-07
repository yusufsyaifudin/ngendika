
// TODO: Replace the following with your app's Firebase project configuration
const firebaseConfig = {
    apiKey: "AIzaSyAoYZnHZxsYg8uOHnpNQHq0iafeT45eOgw",
    authDomain: "gorush-483b6.firebaseapp.com",
    projectId: "gorush-483b6",
    storageBucket: "gorush-483b6.appspot.com",
    messagingSenderId: "965281121928",
    appId: "1:965281121928:web:987ea7946c0fa01635c13d",
    measurementId: "G-RPHDSKL37N"
};

// Initialize Firebase with a default Firebase project
let project = firebase.initializeApp(firebaseConfig);
project.analytics();

// Retrieve Firebase Messaging object.
const messaging = project.messaging();

// Get Instance ID token. Initially this makes a network call, once retrieved
// subsequent calls to getToken will return from cache.
function getToken() {
    messaging.getToken().then((currentToken) => {
        if (currentToken) {
            document.getElementById("token").innerHTML = `Current token: ${currentToken}`;
            console.log('Current token:', currentToken);
        } else {
            // Show permission request.
            document.getElementById("token").innerHTML = 'No Instance ID token available. Request permission to generate one.'
            console.log('No Instance ID token available. Request permission to generate one.');
        }
    }).catch((err) => {
        console.log('An error occurred while retrieving token. ', err);
    });
}

// Callback fired if Instance ID token is updated.
messaging.onTokenRefresh(() => {
    messaging.getToken().then((refreshedToken) => {
        document.getElementById("token").innerHTML = `Refreshed token: ${refreshedToken}`;
        console.log('Token refreshed:', refreshedToken);
        // Indicate that the new Instance ID token has not yet been sent to the
        // app server.
    }).catch((err) => {
        document.getElementById("token").innerHTML = `Unable to retrieve refreshed token: ${err}`;
        console.log('Unable to retrieve refreshed token ', err);
    });
});

// [START receive_message]
// Handle incoming messages. Called when:
// - a message is received while the app has focus
// - the user clicks on an app notification created by a service worker
//   `messaging.setBackgroundMessageHandler` handler.
messaging.onMessage((payload) => {
    document.getElementById("message").innerHTML = JSON.stringify(payload);
    console.log('Message received. ', payload);
});
// [END receive_message]

function requestPermission() {
    console.log('Requesting permission...');
    // [START request_permission]
    Notification.requestPermission().then((permission) => {
        if (permission === 'granted') {
            document.getElementById("log").innerHTML = `Notification permission granted.`;
            console.log('Notification permission granted.');
        } else {
            document.getElementById("log").innerHTML = `Unable to get permission to notify.`;
            console.log('Unable to get permission to notify.', permission);
        }
    });
    // [END request_permission]
}

requestPermission()
getToken()

function deleteToken() {
    // Delete registration token.
    messaging.getToken().then((currentToken) => {
        messaging.deleteToken(currentToken).then(() => {
            console.log('Token deleted.');
            getToken()
        }).catch((err) => {
            console.log('Unable to delete token. ', err);
        });
    }).catch((err) => {
        console.log('Error retrieving registration token. ', err);
    });
}