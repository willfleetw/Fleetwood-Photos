// Import the functions you need from the SDKs you need
import { initializeApp } from 'https://www.gstatic.com/firebasejs/9.9.3/firebase-app.js'
import { getStorage, ref, listAll, getDownloadURL } from 'https://www.gstatic.com/firebasejs/9.9.3/firebase-storage.js'

// TODO: Add SDKs for Firebase products that you want to use
// https://firebase.google.com/docs/web/setup#available-libraries

$(loadGallery());


async function loadGallery() {
  const firebaseConfig = {
    apiKey: "AIzaSyCGvZ6f7efbH0tHfru4SkUuZvdnOHc5LiQ",
    authDomain: "fleetwood-photos.firebaseapp.com",
    projectId: "fleetwood-photos",
    storageBucket: "fleetwood-photos.appspot.com",
    appId: "1:1059550382284:web:b0d1f58561a6ac0d4a9a69"
  };

  // Initialize Firebase
  const app = initializeApp(firebaseConfig);
  const storage = getStorage(app)
  const thumbnailsRef = ref(storage, 'thumbnail')
  var gallery = $('#gallery > .row');
  listAll(thumbnailsRef)
    .then((res) => {
      res.items.forEach((imageRef) => {
        getDownloadURL(imageRef).then((url) => {
          var frame = $('<div>', {
            'class': 'col-4 col-6-medium col-12-small'
          });
          
          var link = $('<a>', {
            href: url,
            'class': 'image fit',
            'data-lightbox': 'general_gallery'
          });

          var thumbnail = $('<img>', {
            src: url
          });

          link.append(thumbnail);
          frame.append(link);
      
          gallery.append(frame);
        })
      })
    });

    $.fn.randomize = function(selector) {
      (selector ? this.find(selector) : this).parent().each(function(){
        $(this).children(selector).sort(function(){
            return Math.random() - 0.5;
        }).detach().appendTo(this);
      });

      return this;
    };

    gallery.randomize();
  
}