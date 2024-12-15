# WebRTC Voice Relay Application

This project demonstrates a WebRTC-based application that connects multiple clients via a signaling server and forwards audio streams through the server (instead of peer-to-peer). The application includes a Go backend for signaling and audio forwarding, and a LitElement-based frontend to interact with WebRTC.

## Features
- Server-forwarded audio streams (not peer-to-peer).
- WebSocket-based signaling for WebRTC session setup.
- Dynamic handling of multiple clients.
- Supports bidirectional audio streams between clients.

---

## Architecture Overview

1. **Frontend (LitElement):**
   - Establishes WebRTC connections using `RTCPeerConnection`.
   - Captures and sends local audio to the server.
   - Receives and plays audio streams forwarded by the server.

2. **Backend (Go):**
   - Manages WebSocket connections for signaling.
   - Handles WebRTC offer/answer exchanges and ICE candidate forwarding.
   - Forwards encrypted RTP packets between clients.

---

## Installation

### Prerequisites
- Go (1.19 or later)
- Node.js (18.x or later) with npm or yarn

### Setup Instructions

1. **Clone the repository:**
   ```bash
   git clone git@github.com:aleksandarjonoski/webrtc-voice-call-app.git
   cd webrtc-voice-call-app
   ```

2. **Install frontend dependencies:**
   ```bash
   cd frontend
   npm install
   npm run build
   ```

3. **Run the backend server:**
   ```bash
   cd ..
   go run main.go
   ```

4. **Access the frontend:**
   Open a browser and navigate to `http://localhost:8080`.

---

## Usage

1. Open the application in two or more browser tabs.
2. Click the "Start Call" button in each tab to join the call.
3. Speak into the microphone to send audio. Each tab will play back the audio streams forwarded by the server.

---

## Detailed Workflow

### Frontend Workflow

1. **WebSocket Connection**:
   - The frontend establishes a WebSocket connection to the backend.
   - Listens for signaling messages from the server (e.g., offers, answers, ICE candidates).

2. **WebRTC Setup**:
   - Creates a `RTCPeerConnection` instance.
   - Captures local audio using `navigator.mediaDevices.getUserMedia`.
   - Sends the local audio track to the server through the WebRTC connection.

3. **Signaling Exchange**:
   - Sends WebRTC offers and ICE candidates to the backend via WebSocket.
   - Processes answers and ICE candidates received from the backend.

4. **Audio Playback**:
   - Receives relayed audio streams and plays them using the `HTMLAudioElement` API.

### Backend Workflow

1. **WebSocket Signaling**:
   - Upgrades HTTP connections to WebSocket for signaling.
   - Maintains a mapping between WebSocket connections and WebRTC peer connections.

2. **WebRTC Session Setup**:
   - Handles SDP offers/answers and ICE candidates.
   - Creates `PeerConnection` instances for each client using the Pion WebRTC library.

3. **Media Forwarding**:
   - Listens for remote tracks from clients.
   - Forwards RTP packets to other connected clients by broadcasting them through their respective peer connections.

---

## Key Components and Functions

### Backend (Go)

1. **WebSocket Endpoint (`HandleWebSocket`)**:
   - Handles WebSocket signaling messages.
   - Manages SDP offers, answers, and ICE candidates.

2. **Peer Connection Management (`CreatePeerConnection`)**:
   - Creates and configures WebRTC peer connections.
   - Adds existing tracks to new peer connections for audio forwarding.

3. **Track Handling**:
   - Receives incoming audio tracks from clients.
   - Broadcasts audio tracks to all other clients.

### Frontend (LitElement)

1. **WebSocket Connection**:
   - Exchanges signaling messages (offers, answers, ICE candidates) with the backend.

2. **WebRTC API Usage**:
   - Captures local audio using `getUserMedia`.
   - Sends audio tracks via `RTCPeerConnection`.
   - Receives and plays remote audio streams.

---

## Example Signaling Flow

1. **Client A starts a call:**
   - Sends an SDP offer to the backend via WebSocket.
   - The backend forwards the offer to other connected clients.

2. **Client B responds:**
   - Sends an SDP answer to the backend.
   - The backend relays the answer back to Client A.

3. **ICE Candidate Exchange:**
   - Both clients exchange ICE candidates through the backend.

4. **Audio Streaming:**
   - Client A sends audio to the backend.
   - The backend forwards the audio to Client B, and vice versa.

---

## Limitations
- **No End-to-End Encryption (E2EE):** Audio is relayed through the server, so it’s not end-to-end encrypted by default.
- **Server Bottleneck:** All audio streams pass through the server, which may limit scalability.

---

## Future Improvements
- Implement end-to-end encryption using WebRTC Insertable Streams API.
- Add video streaming support.
- Improve scalability with a media server like Janus or Jitsi.
- Add support for authentication and secure key exchange.
---

## Acknowledgments
- [Pion WebRTC](https://github.com/pion/webrtc) for the WebRTC Go library.
- [LitElement](https://lit.dev/) for the frontend framework.
- [Gorilla WebSocket](https://github.com/gorilla/websocket) for WebSocket support in Go.

