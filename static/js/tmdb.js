// represent TMDB functionality
function TMDBObject(apiKey) {
    this.apiKey = apiKey;

    // search
    this.search = function(resultFunc, query) {
        var composedRequest = "search/movie?api_key=" + this.apiKey + "&query=" + query + "&include_adult=false&sort_by=popularity.desc";
        this.request(composedRequest, resultFunc);
    };

    // discover
    this.discover = function(resultFunc) {
        var composedRequest = "discover/movie?api_key=" + this.apiKey + "&primary_release_year=" + (new Date()).getFullYear() + "&sort_by=popularity.desc";
        this.request(composedRequest, resultFunc);
    };

    // get film data for specific film IDs
    this.getFilmsByUserID = function(films) {

    };

    // perform the AJAX request
    this.request = function(URL, resultFunc) {
        $.ajax({
            url: "https://api.themoviedb.org/3/" + URL,
            type: 'GET',
            dataType: 'json',
            error: function(e) {
                console.log(e);
            },
            success: function(e) {
                resultFunc(e);
            }
        });
    };
}