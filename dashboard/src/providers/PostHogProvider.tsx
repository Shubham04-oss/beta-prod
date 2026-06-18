'use client'
import posthog from 'posthog-js'
import { PostHogProvider } from 'posthog-js/react'

import * as Sentry from '@sentry/nextjs'

if (typeof window !== 'undefined') {
  posthog.init(process.env.NEXT_PUBLIC_POSTHOG_KEY || 'dummy-key', {
    api_host: process.env.NEXT_PUBLIC_POSTHOG_HOST || 'https://app.posthog.com',
    capture_pageview: false,
    session_recording: {
      maskAllInputs: false,
      maskTextSelector: null,
    }
  })

  // Link PostHog Session ID to Sentry Errors for rapid debugging
  posthog.onSessionId((sessionId) => {
    Sentry.setTag('posthog_session_id', sessionId)
  })
}

export function CSPostHogProvider({ children }: { children: React.ReactNode }) {
  return <PostHogProvider client={posthog}>{children}</PostHogProvider>
}
