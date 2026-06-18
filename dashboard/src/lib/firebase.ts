import { initializeApp, getApps, getApp } from "firebase/app";
import { getAuth, connectAuthEmulator } from "firebase/auth";

const firebaseConfig = {
  projectId: process.env.NEXT_PUBLIC_FIREBASE_PROJECT_ID || "demo-synq",
  apiKey: process.env.NEXT_PUBLIC_FIREBASE_API_KEY || "fake-api-key", // Not verified when using emulator
  authDomain: process.env.NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN || `${process.env.NEXT_PUBLIC_FIREBASE_PROJECT_ID || "demo-synq"}.firebaseapp.com`,
  storageBucket: process.env.NEXT_PUBLIC_FIREBASE_STORAGE_BUCKET,
  messagingSenderId: process.env.NEXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID,
  appId: process.env.NEXT_PUBLIC_FIREBASE_APP_ID,
};

const app = !getApps().length ? initializeApp(firebaseConfig) : getApp();
const auth = getAuth(app);

// Use the remote emulator in development or when explicitly requested
if (typeof window !== "undefined" && (window.location.hostname === "localhost" || process.env.NEXT_PUBLIC_USE_EMULATOR === "true")) {
  // @ts-expect-error - Next.js hot-reload guard
  const isEmulated = auth._isEmulated;
  if (!isEmulated) {
    const emulatorHost = process.env.NEXT_PUBLIC_FIREBASE_AUTH_EMULATOR_HOST || "http://shubhams-mac-mini.local:9099";
    connectAuthEmulator(auth, emulatorHost, { disableWarnings: true });
    // @ts-expect-error - Next.js hot-reload guard
    auth._isEmulated = true;
  }
}

export { app, auth };
