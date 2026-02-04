import * as pulumi from "@pulumi/pulumi";
/**
 * Manages an Azure AD/Entra ID connector in Dex using the Microsoft-specific connector (type: microsoft). This connector provides Microsoft-specific features like group filtering and domain restrictions.
 */
export declare class AzureMicrosoftConnector extends pulumi.CustomResource {
    /**
     * Get an existing AzureMicrosoftConnector resource's state with the given name, ID, and optional extra
     * properties used to qualify the lookup.
     *
     * @param name The _unique_ name of the resulting resource.
     * @param id The _unique_ provider ID of the resource to lookup.
     * @param opts Optional settings to control the behavior of the CustomResource.
     */
    static get(name: string, id: pulumi.Input<pulumi.ID>, opts?: pulumi.CustomResourceOptions): AzureMicrosoftConnector;
    /**
     * Returns true if the given object is an instance of AzureMicrosoftConnector.  This is designed to work even
     * when multiple copies of the Pulumi SDK have been loaded into the same process.
     */
    static isInstance(obj: any): obj is AzureMicrosoftConnector;
    /**
     * Azure AD application (client) ID.
     */
    readonly clientId: pulumi.Output<string>;
    /**
     * Azure AD application client secret.
     */
    readonly clientSecret: pulumi.Output<string>;
    /**
     * Unique identifier for the Azure Microsoft connector.
     */
    readonly connectorId: pulumi.Output<string>;
    /**
     * Name of the claim that contains group memberships (e.g., 'groups'). Used for group-based access control.
     */
    readonly groups: pulumi.Output<string | undefined>;
    /**
     * Human-readable name for the connector, displayed to users during login.
     */
    readonly name: pulumi.Output<string>;
    /**
     * Redirect URI registered in Azure AD. Must match Dex's callback URL.
     */
    readonly redirectUri: pulumi.Output<string>;
    /**
     * Azure AD tenant identifier. Can be 'common' (any Azure AD account), 'organizations' (any organizational account), or a specific tenant ID (UUID format).
     */
    readonly tenant: pulumi.Output<string>;
    /**
     * Create a AzureMicrosoftConnector resource with the given unique name, arguments, and options.
     *
     * @param name The _unique_ name of the resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param opts A bag of options that control this resource's behavior.
     */
    constructor(name: string, args: AzureMicrosoftConnectorArgs, opts?: pulumi.CustomResourceOptions);
}
/**
 * The set of arguments for constructing a AzureMicrosoftConnector resource.
 */
export interface AzureMicrosoftConnectorArgs {
    /**
     * Azure AD application (client) ID.
     */
    clientId: pulumi.Input<string>;
    /**
     * Azure AD application client secret.
     */
    clientSecret: pulumi.Input<string>;
    /**
     * Unique identifier for the Azure Microsoft connector.
     */
    connectorId: pulumi.Input<string>;
    /**
     * Name of the claim that contains group memberships (e.g., 'groups'). Used for group-based access control.
     */
    groups?: pulumi.Input<string>;
    /**
     * Human-readable name for the connector, displayed to users during login.
     */
    name: pulumi.Input<string>;
    /**
     * Redirect URI registered in Azure AD. Must match Dex's callback URL.
     */
    redirectUri: pulumi.Input<string>;
    /**
     * Azure AD tenant identifier. Can be 'common' (any Azure AD account), 'organizations' (any organizational account), or a specific tenant ID (UUID format).
     */
    tenant: pulumi.Input<string>;
}
