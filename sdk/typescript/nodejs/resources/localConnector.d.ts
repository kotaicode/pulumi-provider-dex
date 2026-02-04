import * as pulumi from "@pulumi/pulumi";
/**
 * Manages a local/builtin connector in Dex. The local connector provides username/password authentication stored in Dex's database. This is useful for testing or when you don't have an external identity provider.
 */
export declare class LocalConnector extends pulumi.CustomResource {
    /**
     * Get an existing LocalConnector resource's state with the given name, ID, and optional extra
     * properties used to qualify the lookup.
     *
     * @param name The _unique_ name of the resulting resource.
     * @param id The _unique_ provider ID of the resource to lookup.
     * @param opts Optional settings to control the behavior of the CustomResource.
     */
    static get(name: string, id: pulumi.Input<pulumi.ID>, opts?: pulumi.CustomResourceOptions): LocalConnector;
    /**
     * Returns true if the given object is an instance of LocalConnector.  This is designed to work even
     * when multiple copies of the Pulumi SDK have been loaded into the same process.
     */
    static isInstance(obj: any): obj is LocalConnector;
    /**
     * Unique identifier for the local connector.
     */
    readonly connectorId: pulumi.Output<string>;
    /**
     * Whether the local connector is enabled. Defaults to true.
     */
    readonly enabled: pulumi.Output<boolean | undefined>;
    /**
     * Human-readable name for the connector, displayed to users during login.
     */
    readonly name: pulumi.Output<string>;
    /**
     * Create a LocalConnector resource with the given unique name, arguments, and options.
     *
     * @param name The _unique_ name of the resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param opts A bag of options that control this resource's behavior.
     */
    constructor(name: string, args: LocalConnectorArgs, opts?: pulumi.CustomResourceOptions);
}
/**
 * The set of arguments for constructing a LocalConnector resource.
 */
export interface LocalConnectorArgs {
    /**
     * Unique identifier for the local connector.
     */
    connectorId: pulumi.Input<string>;
    /**
     * Whether the local connector is enabled. Defaults to true.
     */
    enabled?: pulumi.Input<boolean>;
    /**
     * Human-readable name for the connector, displayed to users during login.
     */
    name: pulumi.Input<string>;
}
