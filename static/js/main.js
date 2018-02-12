var thumbTemplate = "";
Mustache.parse(thumbTemplate);


$(document).ready(function() {
    performRequest(hostname + "/static/templates/film_search_result.html", "GET", "", function(result) {
        thumbTemplate = result;
        console.log(thumbTemplate);
    });

    $("#main-panel .row").empty();

    $('#main-nav-search').on('input', function(e){
        if ($(this).val() != "") {
            var query = new TMDBObject("search", $(this).val());
            query.request(populateMainPanel)
        }
    });
});

// represent TMDB functionality
function TMDBObject(operation, query) {
    this.apiKey = "9d35385bd6e30e1da8b4350e5be48b44";
    this.operation = operation;
    this.query = query;

    this.request = function(resultMethod) {
        // prepare ajax request params
        var composedURL = "https://api.themoviedb.org/3/" + this.operation + "/movie?api_key=" + this.apiKey + "&query=" + this.query + "&include_adult=false&sort_by=popularity.desc";

        // perform ajax request
        $.ajax({
            url: composedURL,
            type: 'GET',
            dataType: 'json',
            error: function(e) {
                // resultMethod(e);
                console.log(e);
            },
            success: function(e) {
                resultMethod(e);
            }
        });
    }
}

// populate page with ajax result
function populateMainPanel(content) {
    //$("#main-panel").append(JSON.stringify(content["results"][0], null, "\t"s));

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

    $(".thumbnail-container .hide-btn").on("mousedown", function() {
        $(this).closest(".thumbnail-container").remove();
    });
}