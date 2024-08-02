import { LitElement, html, css } from 'https://cdn.jsdelivr.net/gh/lit/dist@3/core/lit-core.min.js';

// green-500
const validSVG = html`
  <svg width="20" height="20" viewBox="0 0 20 20" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M7.5 13.5L4 10L3 11L7.5 15.5L17.5 5.5L16.5 4.5L7.5 13.5Z" fill="#22c55e"/>
  </svg>
`;

// indigo-600
const loaderSVG = html`
  <svg width="20" height="20" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg">
    <circle cx="10" cy="10" r="8" stroke="#4f46e5" stroke-width="2" fill="none" />
    <circle cx="10" cy="2" r="2" fill="#4f46e5">
      <animateTransform
        attributeName="transform"
        type="rotate"
        from="0 10 10"
        to="360 10 10"
        dur="1s"
        repeatCount="indefinite" />
    </circle>
  </svg>
`;

// red-500
const errorSVG = html`
  <svg width="20" height="20" viewBox="0 0 20 20" fill="none" xmlns="http://www.w3.org/2000/svg">
    <circle cx="10" cy="10" r="9" stroke="#ef4444" stroke-width="2" fill="#ef4444"/>
    <path d="M7 7L13 13M13 7L7 13" stroke="#fff" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
  </svg>
`;

const states = {
  initial: 'initial',
  loading: 'loading',
  valid: 'valid',
  error: 'error'
};

/**
 * @param {string} url
 * @returns {Promise<boolean>}
 */
function validateVideoUrl(url) {
  return new Promise((resolve, reject) => {
    /** @type {HTMLVideoElement} */
    const video = document.createElement('video');

    video.addEventListener('loadedmetadata', () => {
      resolve();
    });

    video.addEventListener('error', () => {
      reject();
    });

    video.src = url;

    video.load();
  });
}

class VideoInput extends LitElement {
  static formAssociated = true;
  static styles = css`
    :host {
      display: block;
    }

    .video-input {
      display: block;
      width: 100%;
      border-radius: 0.375rem;
      border-width: 0;
      padding-top: 0.375rem;
      padding-bottom: 0.375rem;
      padding-right: 2.25rem;
      color: #111827;
      box-shadow: inset 0 0 0 1px rgba(209, 213, 219, 1), 0 1px 2px 0 rgba(0, 0, 0, 0.05);
      font-size: 0.875rem;
      line-height: 1.5rem;
    }

    .video-input::placeholder {
      color: #9CA3AF;
    }

    .video-input:focus {
      box-shadow: inset 0 0 0 2px rgba(99, 102, 241, 1), 0 0 0 2px rgba(99, 102, 241, 1);
    }

    .video-container {
      position: relative;
      margin-top: 0.5rem;
      display: flex;
      align-items: center;
      gap: 0.5rem;
    }

    .indicator {
      position: absolute;
      right: 0.5rem;
    }
  `;

  static properties = {
    id: { type: String },
    name: { type: String },
    state: { type: String },
    value: { type: String, reflect: true },
  }

  constructor() {
    super();
    this.state = states.initial;
    this.id = 'video-url';
    this.name = 'video-url';
    this.value = '';
  }

  updated(changedProperties) {
    if (changedProperties.has('value')) {
      this.handleInput({ target: { value: this.value } });
    }
  }

  isVideoUrlValid() {
    return this.state === states.valid;
  }

  handleInput(e) {
    clearTimeout(this.timeout);
    this.value = e.target.value;
    this.state = states.loading;
    this.requestUpdate();
    this.timeout = setTimeout(async () => {
      const value = this.value;
      if (value !== '') {
        try {
          await validateVideoUrl(value);
          this.state = states.valid;
          const changeEvent = new Event('change', {
            bubbles: true,
            composed: true,
          });
          this.dispatchEvent(changeEvent);
        } catch (error) {
          this.state = states.error;
          this.value = '';
        }
        this.requestUpdate();
      } else {
        this.state = states.error;
        this.value = '';
        this.requestUpdate();
      }
    }, 500);
  }

  render() {
    return html`
      <div class="video-container">
        <input
          id="${this.id}"
          name="${this.name}"
          type="url"
          pattern="https://.*"
          placeholder="https://example.com"
          class="video-input"
          .value="${this.value}"
          required
          @input="${this.handleInput}"
        />
        <div class="indicator">
          ${this.state === states.loading ? loaderSVG : ''}
          ${this.state === states.valid ? validSVG : ''}
          ${this.state === states.error ? errorSVG : ''}
        </div>
      </div>
    `;
  }
}

customElements.define('video-input', VideoInput);
