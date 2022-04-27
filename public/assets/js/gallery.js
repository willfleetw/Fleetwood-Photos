$(loadGallery());

async function loadGallery() {
  var photos = await $.get('data/photos.json');
  var gallery = $('#gallery > .row');

  photos.forEach(photo => {
    var frame = $('<div>', {
      'class': 'col-4 col-6-medium col-12-small'
    });
    
    var link = $('<a>', {
      href: 'images/large/' + photo['name'],
      'class': 'image fit',
      'data-lightbox': 'general_gallery'
    });
    
    var hasTitle = 'title' in photo;
    var hasDesc = 'description' in photo;
    if (hasTitle || hasDesc) {
      var comment = (hasTitle ? photo['title'] : '') 
          + (hasTitle && hasDesc ? ' - ' : '') 
          + (hasDesc ? photo['description'] : '');
      
      link.attr('data-title', comment);
    }
    
    var thumbnail = $('<img>', {
      src: 'images/small/' + photo['name'].replace('Large.jpg', 'Small.jpg')
    });

    link.append(thumbnail);
    frame.append(link);

    gallery.append(frame);
  });
}