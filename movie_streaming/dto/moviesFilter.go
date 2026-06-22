package dto


import (
    "io"
    "strings"
    "errors"
    "encoding/json"
    "github.com/go-playground/validator/v10"

)




type MovieFilters struct{
    Genre   string  `json:"genre" validate:"omitempty,valid_genre"`
    Name    string  `json:"name" validate:"omitempty,required"`
    OrderBy string  `json:"order_by" validate:"omitempty,oneof=desc asc"`
    SortBy  string  `json:"sort_by" validate:"omitempty,oneof=title year rating peers seeds"`
}


type MovieFiltersOMDB struct{
    Name    string     `json:"name" validate:"required"`
    Year    string     `json:"year" validate:"omitempty,valid_year"`
}




var Errors map[string]string  = map[string]string {
    "Genre" : "genre must be one of imdb genres",
    "OrderBy" : "must be oneof='desc asc'",
    "SortBy" : "must be oneof=title year rating peers seeds",
    "Year" : "must be between current year and 1900",
}





func NewMovieFilters(body io.Reader) (*MovieFilters, map[string]string) {
    
    var (
        movieFilters MovieFilters
        field_errors map[string]string
    )


    field_errors = make(map[string]string)
    json.NewDecoder(body).Decode(&movieFilters)
    


    err := Validate.Struct(movieFilters)
    if err != nil {
		var validateErrs validator.ValidationErrors
		if errors.As(err, &validateErrs) {
			for _, e := range validateErrs {
                
                field_errors[strings.ToLower(e.Field())] = Errors[e.Field()]
			}
		}
        return nil, field_errors
    }

    return &movieFilters, nil
}

func NewMovieFiltersOMDB(body io.Reader) (*MovieFiltersOMDB, map[string]string) {
    
    var (
        movieFilters MovieFiltersOMDB
        field_errors map[string]string
    )


    field_errors = make(map[string]string)
    json.NewDecoder(body).Decode(&movieFilters)
    


    err := Validate.Struct(movieFilters)
    if err != nil {
		var validateErrs validator.ValidationErrors
		if errors.As(err, &validateErrs) {
			for _, e := range validateErrs {
                
                field_errors[strings.ToLower(e.Field())] = Errors[e.Field()]
			}
		}
        return nil, field_errors
    }

    return &movieFilters, nil
}


