import { initializeApp } from 'https://www.gstatic.com/firebasejs/9.9.3/firebase-app.js'
import { getDatabase, ref, orderByChild, endBefore, limitToLast, get, query} from 'https://www.gstatic.com/firebasejs/9.9.3/firebase-database.js'

/*
  Firebase Variables
*/
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

/*
  Image Gallery Pagination and Infinite Scrolling Variables
*/
const loader = $('.loader');
const databaseImageCount = (await get(ref(database, "imageCount"))).val();
let cursor = null;
let imageQueryLimit = 0;
let throttleTimer;

let $grid = $('.grid').isotope({
  itemSelector: '.grid-item',
  percentPosition: true,
  masonry: {
    columnWidth: '.grid-sizer',
    gutter: '.gutter-sizer',
  },
  getSortData: {
    priority: '[data-priority] parseInt',
    title: '[data-title]',
    datetaken: '[data-datetaken]',
  },
  sortAscending: {
    priority: false,
    title: true,
    datetaken: true,
  },
});

function filterBy(element, filter) {
  $('.dropdown-item.filter').removeClass('active');
  element.addClass('active');
  $grid.isotope({
    filter: filter,
  });
}
window.filterBy = filterBy;

// Utility Functions
function hideLoader() {
  loader.removeClass('show');
};

function showLoader() {
  loader.addClass('show');
};

function hasMoreImages() {
  return $('.grid-item').length < databaseImageCount;
}

function showImages(images) {
  images.forEach(image => {
    addImageTile(image);
  });
  // I don't know why, but we have to add this event func every time, not just at init
  $grid.imagesLoaded().progress( function() {
    $grid.isotope('layout');
  });
}

const uniqueTags = new Set();
function addToDropdownFilter(tags) {
  let dropdown = $('.dropdown-content')
  tags.forEach(tag => {
    if (!uniqueTags.has(tag)) {
      uniqueTags.add(tag);

      let filter = $('<div>', {
        'class': 'dropdown-item filter',
      });
      filter.on("click", function() {
        filterBy($(this), '.' + tag);
      });
      filter.text(tag.toLocaleUpperCase());

      dropdown.append(filter);
    }
  });
}

function addImageTile(image) {
  let miniURL = 'https://firebasestorage.googleapis.com/v0/b/fleetwood-photos.appspot.com/o/images%2Fmini%2F' + image.name + '.jpg?alt=media'
  let smallURL = 'https://firebasestorage.googleapis.com/v0/b/fleetwood-photos.appspot.com/o/images%2Fsmall%2F' + image.name + '.jpg?alt=media'
  let largeURL = 'https://firebasestorage.googleapis.com/v0/b/fleetwood-photos.appspot.com/o/images%2Flarge%2F' + image.name + '.jpg?alt=media'
  let captionSuffix = " - <a download target='_blank' href='" + smallURL + "'>Small File</a> and <a download target='_blank' href='"+ largeURL + "'>Large File</a>"
  
  addToDropdownFilter(image.meta.tags);

  let tileClass = 'grid-item';
  if (image.meta.height > image.meta.width) {
    tileClass += ' tall';
  } else {
    tileClass += ' wide';
  }
  tileClass += " " + image.meta.tags.join(" ");
  let tile = $('<div>', {
    'class': tileClass,
    'data-priority': image.meta.priority,
    'data-title': image.name,
    'data-datetaken': image.meta.dateTaken,
  });
  let imgWrapper = $('<a>', {
    href: miniURL,
    'data-lightbox': 'gallery',
    'data-title': image.name.replaceAll('_', ' ') + captionSuffix,
  });
  let img = $('<img>', {
    src: miniURL,
    'loading': 'lazy',
  });

  tile.append(imgWrapper.append(img));
  $grid.append(tile).isotope('appended', tile);
};

async function getImages() {
  let dbQuery = query(dbRef, orderByChild('priority'), limitToLast(imageQueryLimit));
  if (cursor != null) {
    dbQuery = query(dbRef, orderByChild('priority'), limitToLast(imageQueryLimit), endBefore(cursor.meta.priority, cursor.name));
  }
  let images = [];
  await get(dbQuery).then(snapshot => {
    snapshot.forEach(child => {
      images.unshift({"name": child.key, "meta": child.val()});
    });
    cursor = images[images.length-1];
  }).catch(error => {
    console.log(error);
  });

  return images;
}

async function loadImages(delay = 600) {
  const throttle = (callback, time) => {
    if (throttleTimer) return;
   
    throttleTimer = true;
   
    setTimeout(() => {
      callback();
      throttleTimer = false;
    }, time);
  };

  showLoader();
  throttle(async () => {
    if (hasMoreImages()) {
      let images = await getImages();
      showImages(images);
    }
    hideLoader();
  }, delay);
}

// Start listening for screen size changes
(() => {
  let mqls = [
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
})();

// Load initial page of images, then begin infinite scrolling
await loadImages(0);
window.addEventListener('scroll', async () => {
  if (window.innerHeight + window.scrollY >= document.body.offsetHeight - 5) {
    await loadImages();
  }
}, {
  passive: true
});