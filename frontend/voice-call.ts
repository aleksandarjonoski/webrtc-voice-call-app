import { LitElement, html, css } from "lit";
import { customElement, state } from "lit/decorators.js";

@customElement("voice-call")
export class VoiceCall extends LitElement {
  @state() private pc: RTCPeerConnection | null = null;
  @state() private ws: WebSocket | null = null;

  static styles = css`
    :host {
      display: block;
      padding: 16px;
    }
  `;

  render() {
    return html` <button @click="${this.startCall}">Start Call</button> `;
  }

  async startCall() {
    this.ws = new WebSocket("ws://localhost:8080/ws");

    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      switch (message.type) {
        case "answer":
          this.handleAnswer(message.payload);
          break;
        case "ice-candidate":
          this.handleRemoteCandidate(message.payload);
          break;
      }
    };

    this.pc = new RTCPeerConnection();

    this.pc.onicecandidate = (event) => {
      if (event.candidate) {
        this.sendMessage("ice-candidate", event.candidate);
      }
    };

    this.pc.ontrack = (event) => {
      const audioElement = document.createElement("audio");
      audioElement.srcObject = event.streams[0];
      audioElement.autoplay = true;
      document.body.appendChild(audioElement);
    };

    const localStream = await navigator.mediaDevices.getUserMedia({
      audio: true,
    });
    localStream.getTracks().forEach((track) => {
      this.pc!.addTrack(track, localStream);
    });

    const offer = await this.pc.createOffer();
    await this.pc.setLocalDescription(offer);

    this.sendMessage("offer", offer);
  }

  async handleAnswer(answer: RTCSessionDescriptionInit) {
    if (this.pc) {
      await this.pc.setRemoteDescription(new RTCSessionDescription(answer));
    }
  }

  handleRemoteCandidate(candidate: RTCIceCandidateInit) {
    if (this.pc) {
      this.pc.addIceCandidate(new RTCIceCandidate(candidate));
    }
  }

  sendMessage(type: string, payload: any) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type, payload }));
    }
  }
}
