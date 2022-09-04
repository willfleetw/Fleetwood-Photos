import { initializeApp } from 'https://www.gstatic.com/firebasejs/9.9.3/firebase-app.js'
import { getDatabase, ref, orderByChild, get, query} from 'https://www.gstatic.com/firebasejs/9.9.3/firebase-database.js'

$(window).on('load', function () {
  const firebaseConfig = {
    apiKey: 'AIzaSyCGvZ6f7efbH0tHfru4SkUuZvdnOHc5LiQ',
    authDomain: 'fleetwood-photos.firebaseapp.com',
    projectId: 'fleetwood-photos',
    databaseURL: 'https://fleetwood-photos-default-rtdb.firebaseio.com',
    appId: '1:1059550382284:web:b0d1f58561a6ac0d4a9a69'
  };
  
  // Initialize Firebase
  const app = initializeApp(firebaseConfig);
  const dbRef = ref(getDatabase(app));
  
  var grid = $('.grid');
  
  get(query(dbRef, orderByChild('priority'))).then((snapshot) => {
    snapshot.forEach((child) => {
      var miniURL = 'https://firebasestorage.googleapis.com/v0/b/fleetwood-photos.appspot.com/o/images%2Fmini%2F' + child.key + '.jpg?alt=media'
      var smallURL = 'https://firebasestorage.googleapis.com/v0/b/fleetwood-photos.appspot.com/o/images%2Fsmall%2F' + child.key + '.jpg?alt=media'
      var largeURL = 'https://firebasestorage.googleapis.com/v0/b/fleetwood-photos.appspot.com/o/images%2Flarge%2F' + child.key + '.jpg?alt=media'
      var captionSuffix = " - <a download target='_blank' href='" + smallURL + "'>Small File</a> and <a download target='_blank' href='"+ largeURL + "'>Large File</a>"
      
      var imgData = child.val();
      var divClass = 'grid-item';
      if (imgData.height > imgData.width) {
        divClass += ' grid-item--height2';
      }
      var item = $('<div>', {
        'class': divClass,
      });
      var imgWrapper = $('<a>', {
        href: miniURL,
        'data-lightbox': 'gallery',
        'data-title': child.key.replaceAll('_', ' ') + captionSuffix,
      });
      var img = $('<img>', {
        src: miniURL,
        'loading': 'lazy',
      });

      imgWrapper.append(img);
      item.append(imgWrapper);
      grid.prepend(item); // prepend since firebase returns ascending order, and we want higher priority shown first
      var $grid = $('.grid').imagesLoaded( function() {
        // init Masonry after all images have loaded
        $grid.masonry({
          itemSelector: '.grid-item',
          // use element for option
          columnWidth: '.grid-sizer',
          gutter: 6,
          percentPosition: true
        });
      });
    });
  }).catch((error) => {
    console.error(error);
  });
});

