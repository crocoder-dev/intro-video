var container = null;
var video = null;

if (!config) {
  var config = {};
}

function calculateWidth(area, aspectRatio) {
  return Math.sqrt(area * aspectRatio);
}

function cleanUp() {
  if (container) {
    container.remove();
    container = null;
    video = null;
  }
}

function loadContainer() {
  if (container) {
    return container
  }

  let unableToFindContainer = false;

  if (config.target !== null) {
    try {
      container = document.querySelector(config.target);
      if (container === null) {
        unableToFindContainer = true;
      } else {
        return container;
      }
    } catch (e) {
      unableToFindContainer = true;
    }
  }

  if (config.target === null || unableToFindContainer) {
    const body = document.querySelector('body');
    container = document.createElement('div');

    body.appendChild(container);
  }

  container.classList.add('iv');

  return container;
}

function preload(videoConfig, callback) {
  container = loadContainer();
  video = document.createElement('video');

  video.addEventListener('loadeddata', () => {
    console.log('loadeddata');
    const ratio = video.videoWidth / video.videoHeight;
    videoConfig.small.width = calculateWidth(284 * 160, ratio);
    videoConfig.small.height = videoConfig.small.width / ratio;

    videoConfig.large.width = calculateWidth(480 * 270, ratio);
    videoConfig.large.height = videoConfig.large.width / ratio;
    if (callback) {
      callback();
    }
  });

  video.classList.add('iv-player');

  video.muted = true;
  video.loop = true;
  video.draggable = false;
  video.src = videoConfig.url;

  container.appendChild(video);
  console.log('t', typeof container, typeof video);
}

function setupIntroVideo({videoConfig, bubble, cta }) {
  const card = document.createElement('div');
  card.classList.add('iv-card');

  card.style.width = `${videoConfig.small.width}px`;
  card.style.height = `${videoConfig.small.height}px`;


  const videoWrapper = document.createElement('div');
  videoWrapper.classList.add('iv-player-wrapper');

  video.style.display = 'block';

  const progressBar = document.createElement('progress');
  progressBar.classList.add('iv-progressbar');
  progressBar.value = 0;
  progressBar.max = 100;

  const button = document.createElement('button');
  button.classList.add('iv-close-button');
  button.innerHTML = '&times;';

  function updateProgressBar() {
    if (!progressBar) return;
    const percentage = (video.currentTime / video.duration) * 100;
    progressBar.value = percentage;
    requestAnimationFrame(updateProgressBar);
  }

  video.addEventListener('play', () => requestAnimationFrame(updateProgressBar))

  button.onclick = () => {
    card.style.opacity = 0;
    setTimeout(() => {
      cleanUp();
    }, 500);
  }

  videoWrapper.onclick = () => {
    card.style.height = `${videoConfig.large.height}px`;
    card.style.width = `${videoConfig.large.width}px`;
    video.muted = false;
    if (cta) {
      videoWrapper.appendChild(cta);
    }
    if (bubble) {
      bubble.remove();
    }
  }

  videoWrapper.appendChild(video);
  videoWrapper.appendChild(progressBar);
  card.appendChild(videoWrapper);
  card.appendChild(button);
  if (bubble) {
    card.appendChild(bubble);
  }
  container.appendChild(card);
  console.log('t', typeof container, typeof video);
  console.log('video.play()');
  video.play();
}
