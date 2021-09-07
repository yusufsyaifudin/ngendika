// Give the service worker access to Firebase Messaging.
// Note that you can only use Firebase Messaging here, other Firebase libraries
// are not available in the service worker.
importScripts('https://www.gstatic.com/firebasejs/8.6.2/firebase-app.js');
importScripts('https://www.gstatic.com/firebasejs/8.6.2/firebase-analytics.js');
importScripts('https://www.gstatic.com/firebasejs/8.6.2/firebase-messaging.js');


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

// Initialize the Firebase app in the service worker by passing in
// your app's Firebase config object.
// https://firebase.google.com/docs/web/setup#config-object
let project = firebase.initializeApp(firebaseConfig);
project.analytics();

// Retrieve an instance of Firebase Messaging so that it can handle background
// messages.
const messaging = project.messaging();

messaging.setBackgroundMessageHandler(function (payload) {
  document.getElementById("message").innerHTML = `[firebase-messaging-sw.js] Received background message: ${payload}`;
  console.log('[firebase-messaging-sw.js] Received background message ', payload);

});


if ('serviceWorker' in navigator) {
  navigator.serviceWorker.register('../firebase-messaging-sw.js')
    .then(function (registration) {
      document.getElementById("log").innerHTML = `Registration successful, scope is: ${registration}`;
      console.log('Registration successful, scope is:', registration.scope);
    }).catch(function (err) {
      document.getElementById("log").innerHTML = `Service worker registration failed, error: ${err}`;
      console.log('Service worker registration failed, error:', err);
    });
}
