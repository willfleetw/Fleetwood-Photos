// Import the functions you need from the SDKs you need
import { initializeApp } from "firebase/app";
import { getAnalytics } from "firebase/analytics";
// TODO: Add SDKs for Firebase products that you want to use
// https://firebase.google.com/docs/web/setup#available-libraries

// Your web app's Firebase configuration
// For Firebase JS SDK v7.20.0 and later, measurementId is optional
const firebaseConfig = {
  apiKey: "AIzaSyCGvZ6f7efbH0tHfru4SkUuZvdnOHc5LiQ",
  authDomain: "fleetwood-photos.firebaseapp.com",
  projectId: "fleetwood-photos",
  storageBucket: "fleetwood-photos.appspot.com",
  messagingSenderId: "1059550382284",
  appId: "1:1059550382284:web:b0d1f58561a6ac0d4a9a69",
  measurementId: "G-5QYQZ52HVX"
};

// Initialize Firebase
const app = initializeApp(firebaseConfig);
const analytics = getAnalytics(app);
