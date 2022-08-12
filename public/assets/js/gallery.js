$(loadGallery());

async function loadGallery() {
  var photos = await $.get('data/photos.json');
  var gallery = $('#gallery > .row');

  photos.forEach(photo => {
    var frame = $('<div>', {
      'class': 'col-4 col-6-medium col-12-small'
    });
    
    var photoPath = 'images/small/' + photo['name']
    var link = $('<a>', {
      href: photoPath,
      'class': 'image fit',
      'data-lightbox': 'general_gallery'
    });
    
    var imageLink = "images/large/" + photo['name'].replace('Small.jpg', 'Large.jpg')
    var comment = '<a href=\'' + imageLink + '\'>Large Version</a>';
    link.attr('data-title', comment);
    
    var thumbnail = $('<img>', {
      src: photoPath
    });

    link.append(thumbnail);
    frame.append(link);

    gallery.append(frame);
  });
}