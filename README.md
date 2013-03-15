GAE Go image optimizer
======================

Go package for automatically optimizing images uploaded to Google App Engine blobstore. This reduces file size of images and, thus, reducing download times and saving you dollars.

Features:
---------
  * Files are converted to JPEG format.
  * Compression rate is changable.
    * (highly compressed) 0 --> 100 (not much compressed)
    * Defaults to 75 (compressed but not visually noticable).
  * Change image dimensions.
    * This value is the largest allowed dimension for the images.
    * 0 = unlimited / no change.
    * Defaults to 0.
  * Leaves other kind of blobs untouched
  * Returnes the same values as blobstore.ParseUploads()


Usage
-----
  ```go
    import "github.com/tomihiltunen/gae-go-image-optimizer"
    
    func urlPathHandler(w http.ResponseWriter, r *http.Request) {
      // Create options
      o := optimg.NewCompressionOptions(r)

      // Set max size
      o.Size = 1600

      // Set quality
      o.quality = 75

      // Get the automatically optimized blobs and other values
      blobs, other, err := optimg.ParseBlobs(o)

      ...
    }
  ```
