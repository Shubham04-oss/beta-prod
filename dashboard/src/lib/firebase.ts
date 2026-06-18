import { initializeApp, getApps, getApp } from "firebase/app";
import { getAuth, connectAuthEmulator } from "firebase/auth";

const firebaseConfig = {
  projectId: "demo-synq",
  apiKey: "fake-api-key", // Not verified when using emulator
  authDomain: "demo-synq.firebaseapp.com",
};

const app = !getApps().length ? initializeApp(firebaseConfig) : getApp();
const auth = getAuth(app);

// Use the remote emulator running on the Mac Mini
if (typeof window !== "undefined" && window.location.hostname === "localhost") {
  // @ts-expect-error - Next.js hot-reload guard
  const isEmulated = auth._isEmulated;
  if (!isEmulated) {
    connectAuthEmulator(auth, "http://shubhams-mac-mini.local:9099", { disableWarnings: true });
    // @ts-expect-error - Next.js hot-reload guard
    auth._isEmulated = true;
  }
}

export { app, auth };
