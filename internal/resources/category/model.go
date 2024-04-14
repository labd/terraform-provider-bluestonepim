package category

import "github.com/hashicorp/terraform-plugin-framework/types"

// CatalogDataSourceModel describes the data source data model.
type Category struct {
	Id       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Number   types.String `tfsdk:"number"`
	ParentId types.String `tfsdk:"parent_id"`
}
