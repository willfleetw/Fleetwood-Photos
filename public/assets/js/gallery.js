import { initializeApp } from 'https://www.gstatic.com/firebasejs/9.9.3/firebase-app.js'
import { getDatabase, ref, orderByChild, endBefore, limitToLast, get, query} from 'https://www.gstatic.com/firebasejs/9.9.3/firebase-database.js'

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
var $grid = $('.grid').masonry({
  itemSelector: '.grid-item',
  // use element for option
  columnWidth: '.grid-sizer',
  gutter: 6,
  percentPosition: true
});
var page = getFirstPage(dbRef);
page = page.then(getNextPage);

function addImageTile(image) {
  var miniURL = 'https://firebasestorage.googleapis.com/v0/b/fleetwood-photos.appspot.com/o/images%2Fmini%2F' + image.name + '.jpg?alt=media'
  var smallURL = 'https://firebasestorage.googleapis.com/v0/b/fleetwood-photos.appspot.com/o/images%2Fsmall%2F' + image.name + '.jpg?alt=media'
  var largeURL = 'https://firebasestorage.googleapis.com/v0/b/fleetwood-photos.appspot.com/o/images%2Flarge%2F' + image.name + '.jpg?alt=media'
  var captionSuffix = " - <a download target='_blank' href='" + smallURL + "'>Small File</a> and <a download target='_blank' href='"+ largeURL + "'>Large File</a>"
  
  var tileClass = 'grid-item';
  if (image.meta.height > image.meta.width) {
    tileClass += ' grid-item--height2';
  }
  var tile = $('<div>', {
    'class': tileClass,
  });
  var imgWrapper = $('<a>', {
    href: miniURL,
    'data-lightbox': 'gallery',
    'data-title': image.name.replaceAll('_', ' ') + captionSuffix,
  });
  var img = $('<img>', {
    src: miniURL,
    'loading': 'lazy',
  });

  tile.append(imgWrapper.append(img))
  $grid.append(tile).masonry('appended', tile).masonry();
};

var cursor = null;
function getFirstPage() {
  return get(query(dbRef, orderByChild('priority'), limitToLast(10))).then(snapshot => {
    var images = [];
    snapshot.forEach(child => {
      images.unshift({"name": child.key, "meta": child.val()});
    });
    cursor = images[images.length-1];

    images.forEach(image => addImageTile(image));
  }).catch(error => {
    console.log(error);
  });
}

function getNextPage() {
  get(query(dbRef, orderByChild('priority'), limitToLast(10), endBefore(cursor.meta.priority, cursor.name))).then(snapshot => {
    var images = [];
    snapshot.forEach(child => {
      images.unshift({"name": child.key, "meta": child.val()});
    });
    cursor = images[images.length-1];

    images.forEach(image => addImageTile(image));
  }).catch(error => {
    console.log(error);
  });
}

