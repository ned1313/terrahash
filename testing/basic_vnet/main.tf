provider "azurerm" {
  features {}
  
}

resource "azurerm_resource_group" "test" {
  name = "terrahash-test"
  location = "East US"
}

module "vnet" {
  source  = "Azure/vnet/azurerm"
  version = "4.1.0"
  
    resource_group_name = azurerm_resource_group.test.name
    use_for_each = true
    vnet_location = azurerm_resource_group.test.location

}

module "internal" {
    source = "./modules"
}