import { initializeApp } from 'https://www.gstatic.com/firebasejs/9.9.3/firebase-app.js'
import { getDatabase, ref, orderByChild, endBefore, limitToLast, get, query} from 'https://www.gstatic.com/firebasejs/9.9.3/firebase-database.js'

// Initialize Firebase
const firebaseConfig = {
  apiKey: 'AIzaSyCGvZ6f7efbH0tHfru4SkUuZvdnOHc5LiQ',
  authDomain: 'fleetwood-photos.firebaseapp.com',
  projectId: 'fleetwood-photos',
  databaseURL: 'https://fleetwood-photos-default-rtdb.firebaseio.com',
  appId: '1:1059550382284:web:b0d1f58561a6ac0d4a9a69'
};
const app = initializeApp(firebaseConfig);
const database = getDatabase(app);
const dbRef = ref(database, '/images/');
const databaseImageCount = await getImageCount();

// Start listening for screen size changes
var imageQueryLimit;
var mqls = [
  window.matchMedia("screen and (max-width: 1199px)"),
  window.matchMedia("screen and (max-width: 991px)"),
  window.matchMedia("screen and (max-width: 767px)")
];
function handleScreenSize(mql) {
  imageQueryLimit = 12;
  if (mqls[0].matches) {
    imageQueryLimit = 10;
  } 
  if (mqls[1].matches) {
    imageQueryLimit = 8;
  } 
  if (mqls[2].matches) {
    imageQueryLimit = 5;
  }
}
for (var i=0; i<mqls.length; i++){
  handleScreenSize(mqls[i]);
  mqls[i].addEventListener('change', handleScreenSize);
}

// Initialize Masonry
var $grid = $('.grid').masonry({
  itemSelector: '.grid-item',
  columnWidth: '.grid-sizer',
  gutter: 6,
  percentPosition: true
});

window.addEventListener('scroll', async () => {
  if (window.innerHeight + window.pageYOffset >= document.body.offsetHeight) {
    throttle(async () => {
      await loadImages();
    }, 1000)
  }
}, {
  passive: true
});

var cursor = null;
await loadImages();

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

var throttleTimer;
 
const throttle = (callback, time) => {
  if (throttleTimer) return;
 
  throttleTimer = true;
 
  setTimeout(() => {
    callback();
    throttleTimer = false;
  }, time);
};

async function loadImages() {
  if ($('.grid-item').length >= databaseImageCount) {
    return; // no more images
  }
  
  var dbQuery = query(dbRef, orderByChild('priority'), limitToLast(imageQueryLimit));
  if (cursor != null) {
    dbQuery = query(dbRef, orderByChild('priority'), limitToLast(imageQueryLimit), endBefore(cursor.meta.priority, cursor.name));
  }
  await get(dbQuery).then(snapshot => {
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

async function getImageCount() {
  var snapshot = await get(ref(database, "imageCount"))
  return snapshot.val()
}