import * as pulumi from "@pulumi/pulumi";
import * as inputs from "../types/input";
import * as outputs from "../types/output";
/**
 * Manages a generic connector (upstream identity provider) in Dex. Use this resource for connectors not covered by specific connector types, or when you need full control over the connector configuration.
 */
export declare class Connector extends pulumi.CustomResource {
    /**
     * Get an existing Connector resource's state with the given name, ID, and optional extra
     * properties used to qualify the lookup.
     *
     * @param name The _unique_ name of the resulting resource.
     * @param id The _unique_ provider ID of the resource to lookup.
     * @param opts Optional settings to control the behavior of the CustomResource.
     */
    static get(name: string, id: pulumi.Input<pulumi.ID>, opts?: pulumi.CustomResourceOptions): Connector;
    /**
     * Returns true if the given object is an instance of Connector.  This is designed to work even
     * when multiple copies of the Pulumi SDK have been loaded into the same process.
     */
    static isInstance(obj: any): obj is Connector;
    /**
     * Unique identifier for the connector.
     */
    readonly connectorId: pulumi.Output<string>;
    /**
     * Human-readable name for the connector, displayed to users during login.
     */
    readonly name: pulumi.Output<string>;
    /**
     * OIDC-specific configuration. Use this for OIDC-based connectors.
     */
    readonly oidcConfig: pulumi.Output<outputs.resources.OIDCConfig | undefined>;
    /**
     * Raw JSON configuration for the connector. Use this for advanced configurations or connector types not directly supported. If provided, this takes precedence over OIDCConfig.
     */
    readonly rawConfig: pulumi.Output<string | undefined>;
    /**
     * Type of connector (e.g., 'oidc', 'saml', 'ldap'). Must match a connector type supported by Dex.
     */
    readonly type: pulumi.Output<string>;
    /**
     * Create a Connector resource with the given unique name, arguments, and options.
     *
     * @param name The _unique_ name of the resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param opts A bag of options that control this resource's behavior.
     */
    constructor(name: string, args: ConnectorArgs, opts?: pulumi.CustomResourceOptions);
}
/**
 * The set of arguments for constructing a Connector resource.
 */
export interface ConnectorArgs {
    /**
     * Unique identifier for the connector.
     */
    connectorId: pulumi.Input<string>;
    /**
     * Human-readable name for the connector, displayed to users during login.
     */
    name: pulumi.Input<string>;
    /**
     * OIDC-specific configuration. Use this for OIDC-based connectors.
     */
    oidcConfig?: pulumi.Input<inputs.resources.OIDCConfigArgs>;
    /**
     * Raw JSON configuration for the connector. Use this for advanced configurations or connector types not directly supported. If provided, this takes precedence over OIDCConfig.
     */
    rawConfig?: pulumi.Input<string>;
    /**
     * Type of connector (e.g., 'oidc', 'saml', 'ldap'). Must match a connector type supported by Dex.
     */
    type: pulumi.Input<string>;
}
