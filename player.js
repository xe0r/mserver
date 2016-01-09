(function(window, document) {
  function toggleVisible(elem) {
    if (elem.style.display=='none') {
      elem.style.display='block';
    } else {
      elem.style.display='none';
    }
  }

  function clamp(v, min, max) {
    if (v < min)
      return min;
    if (v > max)
      return max;
    return v;
  }

  function HomeMediaPlayer(node) {
    this.len = 0;
    this.filename = "";
    this.url = "";

    this.startPos = 0;

    this.node = node;

    this.video = node.querySelector('video');

    this.controls = node.querySelector('.controls');

    this.progress = node.querySelector('.progress');
    this.p1 = this.progress.querySelector('div');

    function isFullscreen() {
      if (node.mozRequestFullScreen) {
        return document.mozFullScreenElement != null;
      } else if (node.webkitRequestFullscreen) {
        return document.webkitFullscreenElement != null;
      } else {
        return document.fullscreenElement != null;
      }
    }

    function updateSize() {
      console.log(this.video.videoWidth, this.video.videoHeight);
      if (!isFullscreen()) {
        var w = this.video.videoWidth;
        var h = this.video.videoHeight;
        if (!w || !h) {
          w = 600;
          h = 300;
        }

        var maxw = 800;
        if (w > maxw) {
          h = parseInt(h * maxw/w);
          w = maxw;
        }
        node.style.width = w+'px';
        node.style.height = h+'px';
      } else {
        node.style.width = '100%';
        node.style.height = '100%';
      }
    }
    updateSize = updateSize.bind(this);

    function updatePos() {
      if (this.len == 0) {
        this.p1.style.width = "0px";
        return;
      }
      var total = this.progress.offsetWidth;
      var time = this.startPos + this.video.currentTime;
      var r = time/this.len;
      //console.log("pos", this.startPos, this.video.currentTime, r);
      var x = parseInt(total*r);
      this.p1.style.width = x + "px";
    }
    updatePos = updatePos.bind(this);

    function onFullscreenChange(evt) {
      updateSize();
      updatePos();
    }

    document.addEventListener('fullscreenchange', onFullscreenChange, false);
    document.addEventListener('mozfullscreenchange', onFullscreenChange, false);
    document.addEventListener('webkitfullscreenchange', onFullscreenChange, false);

    function enableFullscreen() {
      if (node.mozRequestFullScreen) {
        node.mozRequestFullScreen()
      } else if (node.webkitRequestFullscreen) {
        node.webkitRequestFullscreen();
      } else {
        node.requestFullscreen();
      }
    }

    function disableFullscreen() {
      if (node.mozRequestFullScreen) {
        document.mozCancelFullScreen()
      } else if (node.webkitRequestFullscreen) {
        document.webkitCancelFullScreen();
      } else {
        document.cancelFullscreen();
      }
      updateSize();
    }

    function toggleFullscreen() {
      if (isFullscreen()) {
        disableFullscreen();
      } else {
        enableFullscreen();
      }
    }

    function togglePause() {
      if (this.video.paused) {
        this.video.play();
      } else {
        this.video.pause();
      }
    }
    togglePause = togglePause.bind(this);

    this.controls.addEventListener('click', function(evt) {
      togglePause();
    }, false);

    this.controls.querySelector('.fullscreen').addEventListener('click', function(evt) {
      toggleFullscreen();
      evt.stopPropagation();
    }, false);

    this.video.addEventListener('canplay', function(evt) {
      updateSize();
    }, false);
    this.video.addEventListener('timeupdate', function(evt) {
      updatePos();
    }, false);

    var scb = this.controls.querySelector('.sound-control-bar');
    var scb_ptr = scb.querySelector('div');

    function updateVolume() {
      scb_ptr.style.left = (parseInt(this.video.volume*scb.offsetWidth)-3)+"px";
    }
    updateVolume = updateVolume.bind(this);

    this.video.addEventListener('volumechange', updateVolume, false);
    updateVolume();

    scb.addEventListener('click', (function(evt) {
      if (evt.target != scb) {
        return;
      }
      var v = evt.layerX-15;
      var total = scb.offsetWidth;
      var r = v / total;
      this.video.volume = clamp(r, 0.0, 1.0);
      evt.stopPropagation();
    }).bind(this), false);

    scb.addEventListener('wheel', (function(evt){
      this.video.volume = clamp(this.video.volume - evt.deltaY*0.02, 0.0, 1.0);
      evt.stopPropagation();
      evt.preventDefault();
    }).bind(this), false);

    this.progress.addEventListener('click', (function(evt) {
      if (this.len == 0) {
        return;
      }
      console.log(evt);
      var total = this.progress.offsetWidth;
      var r = evt.layerX / total;
      console.log(r);
      this.seek(this.len*r);
      evt.stopPropagation();
    }).bind(this), false);
  }

  HomeMediaPlayer.prototype.seek = function(pos) {
    this.video.pause();
    this.video.src = this.url+parseInt(pos*100);
    console.log("Seeking to", this.video.src);
    this.startPos = pos;
    this.video.play();
  }

  HomeMediaPlayer.prototype.setFilename = function(filename) {
    if (this.xhr) {
      this.xhr.abort();
      this.xhr = null;
    }
    this.video.src = "";

    this.len = 0;
    this.filename = filename;

    this.url = '/video/' + filename + '?ts=';

    var iurl = '/info/' + filename;
    this.xhr = new XMLHttpRequest();
    this.xhr.open('GET', iurl, true)
    this.xhr.onreadystatechange = (function() { 
      if (this.xhr.readyState == XMLHttpRequest.DONE && this.xhr.status == 200) {
        var resp = JSON.parse(this.xhr.response);
        console.log(resp);
        this.len = parseFloat(resp.format.duration);
        console.log("Length:", this.len);
        this.xhr = null;
      }
    }).bind(this);
    this.xhr.send();

    this.seek(0);
  }

  window.HomeMediaPlayer = HomeMediaPlayer;

  window.addEventListener('load', function(e){
    var player = new HomeMediaPlayer(document.getElementById('node'));

    function fileClicked(evt) {
      player.setFilename(evt.target.dataset.filename);
    }

    function dirClicked(evt) {
      var d1 = evt.target.parentNode;
      var content = d1.querySelector('.dircontent');
      if (content) {
        toggleVisible(content);
        return;
      }
      content = document.createElement('div');
      content.className='dircontent';
      d1.appendChild(content);
      fetchDirinfo(evt.target.dataset.dirname, content);
    }

    function fetchDirinfo(dir, elem) {
      var xhr = new XMLHttpRequest();
      xhr.open('GET', "/dirinfo/" + dir, true)
      xhr.onreadystatechange = function() { 
        if (xhr.readyState == XMLHttpRequest.DONE && xhr.status == 200) {
          var resp = JSON.parse(xhr.response);
          console.log(resp);
          for (dn of resp.dirs) {
            var d1 = document.createElement('div');
            var d = document.createElement('div');
            d.className = "dirinfo";
            d.dataset.dirname = dir + "/" + dn;
            d.appendChild(document.createTextNode(dn));
            d1.appendChild(d);
            elem.appendChild(d1);
            d.addEventListener('click', dirClicked, false);
          }
          for (fn of resp.files) {
            var d = document.createElement('div');
            d.className = "fileinfo";
            d.dataset.filename = dir + "/" + fn;
            d.appendChild(document.createTextNode(fn));
            elem.appendChild(d);
            d.addEventListener('click', fileClicked, false);
          }
        }
      }
      xhr.send();
    }
    fetchDirinfo('', document.getElementById('leftbar'));

  }, false);
})(window, document);