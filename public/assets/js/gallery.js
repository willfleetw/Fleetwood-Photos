import { initializeApp } from 'https://www.gstatic.com/firebasejs/9.9.3/firebase-app.js'
import { getDatabase, ref, child, get} from 'https://www.gstatic.com/firebasejs/9.9.3/firebase-database.js'

$(fillGallery)

async function fillGallery() {
  const firebaseConfig = {
    apiKey: 'AIzaSyCGvZ6f7efbH0tHfru4SkUuZvdnOHc5LiQ',
    authDomain: 'fleetwood-photos.firebaseapp.com',
    projectId: 'fleetwood-photos',
    databaseURL: 'https://fleetwood-photos-default-rtdb.firebaseio.com',
    appId: '1:1059550382284:web:b0d1f58561a6ac0d4a9a69'
  };

  // Initialize Firebase
  const app = initializeApp(firebaseConfig);
  const dbRef = ref(getDatabase(app))

  var masonry = $('.masonry');

  get(dbRef).then((snapshot) => {
    for (const [key, value] of Object.entries(snapshot.val())) {
      var tile = $('<div>', {
        'class': 'mItem',
      });
      var miniURL = 'https://firebasestorage.googleapis.com/v0/b/fleetwood-photos.appspot.com/o/images%2Fmini%2F' + key + '.jpg?alt=media'
      var smallURL = 'https://firebasestorage.googleapis.com/v0/b/fleetwood-photos.appspot.com/o/images%2Fsmall%2F' + key + '.jpg?alt=media'
      var largeURL = 'https://firebasestorage.googleapis.com/v0/b/fleetwood-photos.appspot.com/o/images%2Flarge%2F' + key + '.jpg?alt=media'
      var captionSuffix = " - <a href='" + smallURL + "'>Small File</a> and <a href='"+ largeURL + "'>Large File</a>"

      var lbImg = $('<a>', {
        href: miniURL,
        'data-lightbox': 'gallery',
        'data-title': key.replaceAll('_', ' ') + captionSuffix,
      });
      var img = $('<img>', {
        src: miniURL,
        'loading': 'lazy',
      });

      lbImg.append(img);
      tile.append(lbImg);
      masonry.append(tile);
    }
  }).catch((error) => {
    console.error(error);
  });
}

