"use client";

import { useEffect, useState } from "react";
import { motion, AnimatePresence } from "framer-motion";

export function TimeSyncedBackground() {
  const [bgImage, setBgImage] = useState<string>("");

  useEffect(() => {
    const updateBackground = () => {
      const hour = new Date().getHours();

      // Cycle the scenic backgrounds based on time
      if (hour >= 5 && hour < 9) {
        setBgImage("/dawn.svg");
      } else if (hour >= 9 && hour < 17) {
        setBgImage("/morning.svg");
      } else if (hour >= 17 && hour < 20) {
        setBgImage("/afternoon.svg");
      } else {
        setBgImage("/night.svg");
      }

      // Enforce the white/light theme always
      document.documentElement.classList.remove("dark");
    };

    updateBackground();
    const interval = setInterval(updateBackground, 60000);
    return () => clearInterval(interval);
  }, []);

  if (!bgImage) return null;

  return (
    <div className="fixed inset-0 -z-20 w-full h-full overflow-hidden bg-[#0a0a0a]">
      <AnimatePresence mode="wait">
        <motion.img
          key={bgImage}
          src={bgImage}
          alt="Dashboard background"
          initial={{ opacity: 0, scale: 1.05 }}
          animate={{ opacity: 1, scale: 1 }}
          exit={{ opacity: 0, scale: 0.95 }}
          transition={{ duration: 2, ease: "easeInOut" }}
          className="w-full h-full object-cover"
        />
      </AnimatePresence>
    </div>
  );
}
