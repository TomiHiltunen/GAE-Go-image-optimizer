/***************************************************************
*
*   GAE Go automatic blob image optimizer
*   
*   Created by Tomi Hiltunen 2013.
*   http://www.linkedin.com/in/tomihiltunen
*
*   https://github.com/TomiHiltunen/GAE-Go-image-optimizer
*
*       - Use this script however you wish.
*       - Do not remove any copyrights/comments on any files included.
*       - All use is on your own risk.
*
*   Intented use:
*       - Drop-in replacement for GAE's blobstore.ParseUploads(...)
*       - Automatically optimized any images uploaded
*         to Google App Engine blobstore.
*           - Reduces data amount in the blobstore.
*           - Reduces download times.
*
***************************************************************/
package slugify


import (
    // Go packages
    "strings"
    "regexp"
)


/*
 *  Allowed mime-types.
 *  These should be the ones supported by Go image package.
 */
var (
    allowedMimeTypes = map[string]bool {
        "image/jpeg": true,
        "image/jpg": true,
        "image/png": true,
        "image/gif": true,
    }
)


/*
 * The options for image optimization.
 *
 *      Quality     The quality of the JPEG output (0-100)
 *      Size        Maximum dimension (width/height) for the photo
 *      Request     The pointer for the HTTP request
 *      Context     App Engine context    
 */
type compressionOptions struct {
    Quality     int
    Size        int
    Request     *http.Request
    Context     appengine.Context
}


/*
 * Create new set of options.
 *
 *      - Sets Quality to 75 as default. 75 is highly compressed but not visually noticable.
 *      - Sets Size to 0 which means that no changes to images dimensions will be made.
 *      - Creates new App Engine context.
 */
func DashedSlug(original string) (edited string) {
    // Small caps
    edited = strings.ToLower(original)
    // Trim leading and trailing white-space
    edited = strings.TrimSpace()
    // RegExp
    re, _ := regexp.Compile(`Hello`)


    // Convert spaces to dashes
    re, _ := regexp.Compile(`\s`)
    edited = re.ReplaceAllString(edited, "-")
}


/*
 * This one does the magic.
 *
 *      - Gets the uploaded blobs by calling blobstore.ParseUpload()
 *      - Maintains all other values that come from blobstore.
 *      - Hands out the results for further processing.
 */
func ParseBlobs(options *compressionOptions) (blobs map[string][]*blobstore.BlobInfo, other url.Values, err error) {
    blobs, other, err = blobstore.ParseUpload(options.Request)
    if err != nil {
        return
    }
    // Loop through all the blob names
    for keyName, blobArray := range blobs {
        blobs[keyName] = handleBlobArray(options, blobArray)
    }
    return
}


/*
 * Handles blob arrays and returns the replaced set of blobs.
 */
func handleBlobArray(options *compressionOptions, blobArrayOriginal []*blobstore.BlobInfo) (blobArray []*blobstore.BlobInfo) {
    blobArray = blobArrayOriginal
    // Loop through all the blobs in the array
    for index, blobInfo := range blobArray {
        blobArray[index] = handleBlob(options, blobInfo)
    }
    return
}


/*
 * Handles individual blobs.
 *
 *      - Only supported image types will be processed. Others will be returned as-is.
 *      - Resizes the image if necessary.
 *      - Writes the new compressed JPEG to blobstore.
 *      - Deletes the old blob and substitutes the old BlobInfo with the new one.
 */
func handleBlob(options *compressionOptions, blobOriginal *blobstore.BlobInfo) (blob *blobstore.BlobInfo) {
    blob = blobOriginal
    // Check that the blob is of supported mime-type
    if !validateMimeType(blob) {
        return
    }
    // Instantiate blobstore reader
    reader := blobstore.NewReader(options.Context, blob.BlobKey)
    // Instantiate the image object
    img, _, err := image.Decode(reader)
    if err != nil {
        return
    }
    // Resize if necessary
    // Maintain aspect ratio!
    if options.Size > 0 && (img.Bounds().Max.X > options.Size || img.Bounds().Max.Y > options.Size) {
        size_x := img.Bounds().Max.X
        size_y := img.Bounds().Max.Y
        if size_x > options.Size {
            size_x_before := size_x
            size_x = options.Size
            size_y = int(  math.Floor(  float64(size_y) * float64(float64(size_x)/float64(size_x_before))  )  )
        }
        if size_y > options.Size {
            size_y_before := size_y
            size_y = options.Size
            size_x = int(  math.Floor(  float64(size_x) * float64(float64(size_y)/float64(size_y_before))  )  )
        }
        img = resize.Resize(img, img.Bounds(), size_x, size_y)
    }
    // JPEG options
    o := &jpeg.Options {
        Quality: options.Quality,
    }
    // Open writer
    writer, err := blobstore.Create(options.Context, "image/jpeg")
    if err != nil {
        return
    }
    // Write to blobstore
    if err := jpeg.Encode(writer, img, o); err != nil {
        return
    }
    // Close writer
    if err := writer.Close(); err != nil {
        return
    }
    // Get key
    newKey, err := writer.Key()
    if err != nil {
        return
    }
    // Get new BlobInfo
    newBlobInfo, err := blobstore.Stat(options.Context, newKey)
    if err != nil {
        return
    }
    // All good!
    // Now replace the old blob and delete it
    deleteOldBlob(options, blob.BlobKey)
    blob = newBlobInfo
    return
}


// Validates blob mime-type
func validateMimeType(blob *blobstore.BlobInfo) (bool) {
    mimeType := strings.ToLower(blob.ContentType)
    if match := allowedMimeTypes[mimeType]; match {
        return true
    }
    return false
}


// Removes the old blob from blobstore
func deleteOldBlob(options *compressionOptions, blobkey appengine.BlobKey) {
    _ = blobstore.Delete(options.Context, blobkey)
}