var thumbTemplate = "";
Mustache.parse(thumbTemplate);

var tmdb = new TMDBObject("9d35385bd6e30e1da8b4350e5be48b44");


$(document).ready(function() {
    performRequest(hostname + "/static/templates/film_search_result.html", "GET", "", function(result) {
        thumbTemplate = result;
    });

    $("#main-panel .row").empty();
    tmdb.discover(populateMainPanel);

    $('#main-nav-search').on('input', function() {
        if ($(this).val() !== "") {
            tmdb.search(populateMainPanel, $(this).val())
        }
        else {
            tmdb.discover(populateMainPanel)
        }
    });
});

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
    }
}

// populate page with ajax result
function populateMainPanel(content) {
    // clear main panel
    $("#main-panel .row").empty();

    // iterate over search results
    $.each(content["results"], function (key, value) {
        var film = content["results"][key];
        var imagePath = "http://via.placeholder.com/185x278?text=?";
        if (film["poster_path"] != null) {
            imagePath = "https://image.tmdb.org/t/p/w185" + film["poster_path"];
        }
        var overviewTrimmed = film["overview"];
        var thumbRendered = Mustache.render(thumbTemplate, {title: film["title"], overview: overviewTrimmed, film_image: imagePath});
        $("#main-panel .row").append(thumbRendered);
    });

    $(".thumbnail-container .hide-btn").on("click", function(e) {
        e.preventDefault();
        $(this).closest(".thumbnail-container").remove();
    });

    $(".thumbnail-container .watched-btn").on("click", function() {
        e.preventDefault();
        addToWatchedList($(this).attr("data-file-id"));
    });
}

// Add selected film to current user's watched list
function addToWatchedList(filmID) {
    // get rating for film
    var rating = "5";
    do {
        rating = prompt("Rate the film between 0 and 10...", "5");
        var valid = (parseInt(rating) >= 0 && parseInt(rating) <= 10)
    } while(!valid);

    // get user_id from dropdown


    var data = "user_id=" + + "&film_id=" + filmID + "&rating=" + rating;
    performRequest(hostname + "/watched", "POST", data, function(result) {
        console.log(result);
    });
}