//
// This file is auto generated by modelgen tool
// Do not edit this file manually

var AppProfileSummaryView = React.createClass({
  	render: function() {
		var self = this

		// Walk thru all objects
		var appProfileListView = self.props.appProfiles.map(function(appProfile){
			return (
				<ModalTrigger modal={<AppProfileModalView appProfile={ appProfile }/>}>
					<tr key={ appProfile.key } className="info">
						
						   
					</tr>
				</ModalTrigger>
			);
		});

		return (
        <div>
			<Table hover>
				<thead>
					<tr>
					
					   
					</tr>
				</thead>
				<tbody>
            		{ appProfileListView }
				</tbody>
			</Table>
        </div>
    	);
	}
});

var AppProfileModalView = React.createClass({
	render() {
		var obj = this.props.appProfile
	    return (
	      <Modal {...this.props} bsStyle='primary' bsSize='large' title='AppProfile' animation={false}>
	        <div className='modal-body' style={ {margin: '5%',} }>
			
			
				<Input type='text' label='Application Profile Name' ref='appProfileName' defaultValue={obj.appProfileName} placeholder='Application Profile Name' />
			
				<Input type='text' label='Member groups of the appProf' ref='endpointGroups' defaultValue={obj.endpointGroups} placeholder='Member groups of the appProf' />
			
				<Input type='text' label='Tenant Name' ref='tenantName' defaultValue={obj.tenantName} placeholder='Tenant Name' />
			
			</div>
	        <div className='modal-footer'>
				<Button onClick={this.props.onRequestHide}>Close</Button>
	        </div>
	      </Modal>
	    );
  	}
});

module.exports.AppProfileSummaryView = AppProfileSummaryView
module.exports.AppProfileModalView = AppProfileModalView
var BgpSummaryView = React.createClass({
  	render: function() {
		var self = this

		// Walk thru all objects
		var BgpListView = self.props.Bgps.map(function(Bgp){
			return (
				<ModalTrigger modal={<BgpModalView Bgp={ Bgp }/>}>
					<tr key={ Bgp.key } className="info">
						
						     
					</tr>
				</ModalTrigger>
			);
		});

		return (
        <div>
			<Table hover>
				<thead>
					<tr>
					
					     
					</tr>
				</thead>
				<tbody>
            		{ BgpListView }
				</tbody>
			</Table>
        </div>
    	);
	}
});

var BgpModalView = React.createClass({
	render() {
		var obj = this.props.Bgp
	    return (
	      <Modal {...this.props} bsStyle='primary' bsSize='large' title='Bgp' animation={false}>
	        <div className='modal-body' style={ {margin: '5%',} }>
			
			
				<Input type='text' label='AS id' ref='as' defaultValue={obj.as} placeholder='AS id' />
			
				<Input type='text' label='host name' ref='hostname' defaultValue={obj.hostname} placeholder='host name' />
			
				<Input type='text' label='Bgp  neighbor' ref='neighbor' defaultValue={obj.neighbor} placeholder='Bgp  neighbor' />
			
				<Input type='text' label='AS id' ref='neighbor-as' defaultValue={obj.neighbor-as} placeholder='AS id' />
			
				<Input type='text' label='Bgp router intf ip' ref='routerip' defaultValue={obj.routerip} placeholder='Bgp router intf ip' />
			
			</div>
	        <div className='modal-footer'>
				<Button onClick={this.props.onRequestHide}>Close</Button>
	        </div>
	      </Modal>
	    );
  	}
});

module.exports.BgpSummaryView = BgpSummaryView
module.exports.BgpModalView = BgpModalView
var EndpointSummaryView = React.createClass({
  	render: function() {
		var self = this

		// Walk thru all objects
		var endpointListView = self.props.endpoints.map(function(endpoint){
			return (
				<ModalTrigger modal={<EndpointModalView endpoint={ endpoint }/>}>
					<tr key={ endpoint.key } className="info">
						
						
					</tr>
				</ModalTrigger>
			);
		});

		return (
        <div>
			<Table hover>
				<thead>
					<tr>
					
					
					</tr>
				</thead>
				<tbody>
            		{ endpointListView }
				</tbody>
			</Table>
        </div>
    	);
	}
});

var EndpointModalView = React.createClass({
	render() {
		var obj = this.props.endpoint
	    return (
	      <Modal {...this.props} bsStyle='primary' bsSize='large' title='Endpoint' animation={false}>
	        <div className='modal-body' style={ {margin: '5%',} }>
			
			
			</div>
	        <div className='modal-footer'>
				<Button onClick={this.props.onRequestHide}>Close</Button>
	        </div>
	      </Modal>
	    );
  	}
});

module.exports.EndpointSummaryView = EndpointSummaryView
module.exports.EndpointModalView = EndpointModalView
var EndpointGroupSummaryView = React.createClass({
  	render: function() {
		var self = this

		// Walk thru all objects
		var endpointGroupListView = self.props.endpointGroups.map(function(endpointGroup){
			return (
				<ModalTrigger modal={<EndpointGroupModalView endpointGroup={ endpointGroup }/>}>
					<tr key={ endpointGroup.key } className="info">
						
					</tr>
				</ModalTrigger>
			);
		});

		return (
        <div>
			<Table hover>
				<thead>
					<tr>
					
					</tr>
				</thead>
				<tbody>
            		{ endpointGroupListView }
				</tbody>
			</Table>
        </div>
    	);
	}
});

var EndpointGroupModalView = React.createClass({
	render() {
		var obj = this.props.endpointGroup
	    return (
	      <Modal {...this.props} bsStyle='primary' bsSize='large' title='EndpointGroup' animation={false}>
	        <div className='modal-body' style={ {margin: '5%',} }>
			
			
				<Input type='text' label='External contracts' ref='extContractsGrps' defaultValue={obj.extContractsGrps} placeholder='External contracts' />
			
				<Input type='text' label='Group name' ref='groupName' defaultValue={obj.groupName} placeholder='Group name' />
			
				<Input type='text' label='Network profile name' ref='netProfile' defaultValue={obj.netProfile} placeholder='Network profile name' />
			
				<Input type='text' label='Network' ref='networkName' defaultValue={obj.networkName} placeholder='Network' />
			
				<Input type='text' label='Policies' ref='policies' defaultValue={obj.policies} placeholder='Policies' />
			
				<Input type='text' label='Tenant' ref='tenantName' defaultValue={obj.tenantName} placeholder='Tenant' />
			
			</div>
	        <div className='modal-footer'>
				<Button onClick={this.props.onRequestHide}>Close</Button>
	        </div>
	      </Modal>
	    );
  	}
});

module.exports.EndpointGroupSummaryView = EndpointGroupSummaryView
module.exports.EndpointGroupModalView = EndpointGroupModalView
var ExtContractsGroupSummaryView = React.createClass({
  	render: function() {
		var self = this

		// Walk thru all objects
		var extContractsGroupListView = self.props.extContractsGroups.map(function(extContractsGroup){
			return (
				<ModalTrigger modal={<ExtContractsGroupModalView extContractsGroup={ extContractsGroup }/>}>
					<tr key={ extContractsGroup.key } className="info">
						
					</tr>
				</ModalTrigger>
			);
		});

		return (
        <div>
			<Table hover>
				<thead>
					<tr>
					
					</tr>
				</thead>
				<tbody>
            		{ extContractsGroupListView }
				</tbody>
			</Table>
        </div>
    	);
	}
});

var ExtContractsGroupModalView = React.createClass({
	render() {
		var obj = this.props.extContractsGroup
	    return (
	      <Modal {...this.props} bsStyle='primary' bsSize='large' title='ExtContractsGroup' animation={false}>
	        <div className='modal-body' style={ {margin: '5%',} }>
			
			
				<Input type='text' label='Contracts list' ref='contracts' defaultValue={obj.contracts} placeholder='Contracts list' />
			
				<Input type='text' label='Contracts group name' ref='contractsGroupName' defaultValue={obj.contractsGroupName} placeholder='Contracts group name' />
			
				<Input type='text' label='Contracts type' ref='contractsType' defaultValue={obj.contractsType} placeholder='Contracts type' />
			
				<Input type='text' label='Tenant name' ref='tenantName' defaultValue={obj.tenantName} placeholder='Tenant name' />
			
			</div>
	        <div className='modal-footer'>
				<Button onClick={this.props.onRequestHide}>Close</Button>
	        </div>
	      </Modal>
	    );
  	}
});

module.exports.ExtContractsGroupSummaryView = ExtContractsGroupSummaryView
module.exports.ExtContractsGroupModalView = ExtContractsGroupModalView
var GlobalSummaryView = React.createClass({
  	render: function() {
		var self = this

		// Walk thru all objects
		var globalListView = self.props.globals.map(function(global){
			return (
				<ModalTrigger modal={<GlobalModalView global={ global }/>}>
					<tr key={ global.key } className="info">
						
						    
					</tr>
				</ModalTrigger>
			);
		});

		return (
        <div>
			<Table hover>
				<thead>
					<tr>
					
					    
					</tr>
				</thead>
				<tbody>
            		{ globalListView }
				</tbody>
			</Table>
        </div>
    	);
	}
});

var GlobalModalView = React.createClass({
	render() {
		var obj = this.props.global
	    return (
	      <Modal {...this.props} bsStyle='primary' bsSize='large' title='Global' animation={false}>
	        <div className='modal-body' style={ {margin: '5%',} }>
			
			
				<Input type='text' label='name of this block(must be 'global')' ref='name' defaultValue={obj.name} placeholder='name of this block(must be 'global')' />
			
				<Input type='text' label='Network infrastructure type' ref='networkInfraType' defaultValue={obj.networkInfraType} placeholder='Network infrastructure type' />
			
				<Input type='text' label='Allowed vlan range' ref='vlans' defaultValue={obj.vlans} placeholder='Allowed vlan range' />
			
				<Input type='text' label='Allwed vxlan range' ref='vxlans' defaultValue={obj.vxlans} placeholder='Allwed vxlan range' />
			
			</div>
	        <div className='modal-footer'>
				<Button onClick={this.props.onRequestHide}>Close</Button>
	        </div>
	      </Modal>
	    );
  	}
});

module.exports.GlobalSummaryView = GlobalSummaryView
module.exports.GlobalModalView = GlobalModalView
var NetprofileSummaryView = React.createClass({
  	render: function() {
		var self = this

		// Walk thru all objects
		var netprofileListView = self.props.netprofiles.map(function(netprofile){
			return (
				<ModalTrigger modal={<NetprofileModalView netprofile={ netprofile }/>}>
					<tr key={ netprofile.key } className="info">
						
						    
					</tr>
				</ModalTrigger>
			);
		});

		return (
        <div>
			<Table hover>
				<thead>
					<tr>
					
					    
					</tr>
				</thead>
				<tbody>
            		{ netprofileListView }
				</tbody>
			</Table>
        </div>
    	);
	}
});

var NetprofileModalView = React.createClass({
	render() {
		var obj = this.props.netprofile
	    return (
	      <Modal {...this.props} bsStyle='primary' bsSize='large' title='Netprofile' animation={false}>
	        <div className='modal-body' style={ {margin: '5%',} }>
			
			
				<Input type='text' label='DSCP' ref='DSCP' defaultValue={obj.DSCP} placeholder='DSCP' />
			
				<Input type='text' label='Allocated bandwidth' ref='bandwidth' defaultValue={obj.bandwidth} placeholder='Allocated bandwidth' />
			
				<Input type='text' label='Network profile name' ref='profileName' defaultValue={obj.profileName} placeholder='Network profile name' />
			
				<Input type='text' label='Tenant name' ref='tenantName' defaultValue={obj.tenantName} placeholder='Tenant name' />
			
			</div>
	        <div className='modal-footer'>
				<Button onClick={this.props.onRequestHide}>Close</Button>
	        </div>
	      </Modal>
	    );
  	}
});

module.exports.NetprofileSummaryView = NetprofileSummaryView
module.exports.NetprofileModalView = NetprofileModalView
var NetworkSummaryView = React.createClass({
  	render: function() {
		var self = this

		// Walk thru all objects
		var networkListView = self.props.networks.map(function(network){
			return (
				<ModalTrigger modal={<NetworkModalView network={ network }/>}>
					<tr key={ network.key } className="info">
						
						         
					</tr>
				</ModalTrigger>
			);
		});

		return (
        <div>
			<Table hover>
				<thead>
					<tr>
					
					         
					</tr>
				</thead>
				<tbody>
            		{ networkListView }
				</tbody>
			</Table>
        </div>
    	);
	}
});

var NetworkModalView = React.createClass({
	render() {
		var obj = this.props.network
	    return (
	      <Modal {...this.props} bsStyle='primary' bsSize='large' title='Network' animation={false}>
	        <div className='modal-body' style={ {margin: '5%',} }>
			
			
				<Input type='text' label='Encapsulation' ref='encap' defaultValue={obj.encap} placeholder='Encapsulation' />
			
				<Input type='text' label='Gateway' ref='gateway' defaultValue={obj.gateway} placeholder='Gateway' />
			
				<Input type='text' label='IPv6Gateway' ref='ipv6Gateway' defaultValue={obj.ipv6Gateway} placeholder='IPv6Gateway' />
			
				<Input type='text' label='IPv6Subnet' ref='ipv6Subnet' defaultValue={obj.ipv6Subnet} placeholder='IPv6Subnet' />
			
				<Input type='text' label='Network name' ref='networkName' defaultValue={obj.networkName} placeholder='Network name' />
			
				<Input type='text' label='Network Type' ref='nwType' defaultValue={obj.nwType} placeholder='Network Type' />
			
				<Input type='text' label='Vlan/Vxlan Tag' ref='pktTag' defaultValue={obj.pktTag} placeholder='Vlan/Vxlan Tag' />
			
				<Input type='text' label='Subnet' ref='subnet' defaultValue={obj.subnet} placeholder='Subnet' />
			
				<Input type='text' label='Tenant Name' ref='tenantName' defaultValue={obj.tenantName} placeholder='Tenant Name' />
			
			</div>
	        <div className='modal-footer'>
				<Button onClick={this.props.onRequestHide}>Close</Button>
	        </div>
	      </Modal>
	    );
  	}
});

module.exports.NetworkSummaryView = NetworkSummaryView
module.exports.NetworkModalView = NetworkModalView
var PolicySummaryView = React.createClass({
  	render: function() {
		var self = this

		// Walk thru all objects
		var policyListView = self.props.policys.map(function(policy){
			return (
				<ModalTrigger modal={<PolicyModalView policy={ policy }/>}>
					<tr key={ policy.key } className="info">
						
						  
					</tr>
				</ModalTrigger>
			);
		});

		return (
        <div>
			<Table hover>
				<thead>
					<tr>
					
					  
					</tr>
				</thead>
				<tbody>
            		{ policyListView }
				</tbody>
			</Table>
        </div>
    	);
	}
});

var PolicyModalView = React.createClass({
	render() {
		var obj = this.props.policy
	    return (
	      <Modal {...this.props} bsStyle='primary' bsSize='large' title='Policy' animation={false}>
	        <div className='modal-body' style={ {margin: '5%',} }>
			
			
				<Input type='text' label='Policy Name' ref='policyName' defaultValue={obj.policyName} placeholder='Policy Name' />
			
				<Input type='text' label='Tenant Name' ref='tenantName' defaultValue={obj.tenantName} placeholder='Tenant Name' />
			
			</div>
	        <div className='modal-footer'>
				<Button onClick={this.props.onRequestHide}>Close</Button>
	        </div>
	      </Modal>
	    );
  	}
});

module.exports.PolicySummaryView = PolicySummaryView
module.exports.PolicyModalView = PolicyModalView
var RuleSummaryView = React.createClass({
  	render: function() {
		var self = this

		// Walk thru all objects
		var ruleListView = self.props.rules.map(function(rule){
			return (
				<ModalTrigger modal={<RuleModalView rule={ rule }/>}>
					<tr key={ rule.key } className="info">
						
						              
					</tr>
				</ModalTrigger>
			);
		});

		return (
        <div>
			<Table hover>
				<thead>
					<tr>
					
					              
					</tr>
				</thead>
				<tbody>
            		{ ruleListView }
				</tbody>
			</Table>
        </div>
    	);
	}
});

var RuleModalView = React.createClass({
	render() {
		var obj = this.props.rule
	    return (
	      <Modal {...this.props} bsStyle='primary' bsSize='large' title='Rule' animation={false}>
	        <div className='modal-body' style={ {margin: '5%',} }>
			
			
				<Input type='text' label='Action' ref='action' defaultValue={obj.action} placeholder='Action' />
			
				<Input type='text' label='Direction' ref='direction' defaultValue={obj.direction} placeholder='Direction' />
			
				<Input type='text' label='From Endpoint Group' ref='fromEndpointGroup' defaultValue={obj.fromEndpointGroup} placeholder='From Endpoint Group' />
			
				<Input type='text' label='IP Address' ref='fromIpAddress' defaultValue={obj.fromIpAddress} placeholder='IP Address' />
			
				<Input type='text' label='From Network' ref='fromNetwork' defaultValue={obj.fromNetwork} placeholder='From Network' />
			
				<Input type='text' label='Policy Name' ref='policyName' defaultValue={obj.policyName} placeholder='Policy Name' />
			
				<Input type='text' label='Port No' ref='port' defaultValue={obj.port} placeholder='Port No' />
			
				<Input type='text' label='Priority' ref='priority' defaultValue={obj.priority} placeholder='Priority' />
			
				<Input type='text' label='Protocol' ref='protocol' defaultValue={obj.protocol} placeholder='Protocol' />
			
				<Input type='text' label='Rule Id' ref='ruleId' defaultValue={obj.ruleId} placeholder='Rule Id' />
			
				<Input type='text' label='Tenant Name' ref='tenantName' defaultValue={obj.tenantName} placeholder='Tenant Name' />
			
				<Input type='text' label='To Endpoint Group' ref='toEndpointGroup' defaultValue={obj.toEndpointGroup} placeholder='To Endpoint Group' />
			
				<Input type='text' label='IP Address' ref='toIpAddress' defaultValue={obj.toIpAddress} placeholder='IP Address' />
			
				<Input type='text' label='To Network' ref='toNetwork' defaultValue={obj.toNetwork} placeholder='To Network' />
			
			</div>
	        <div className='modal-footer'>
				<Button onClick={this.props.onRequestHide}>Close</Button>
	        </div>
	      </Modal>
	    );
  	}
});

module.exports.RuleSummaryView = RuleSummaryView
module.exports.RuleModalView = RuleModalView
var ServiceLBSummaryView = React.createClass({
  	render: function() {
		var self = this

		// Walk thru all objects
		var serviceLBListView = self.props.serviceLBs.map(function(serviceLB){
			return (
				<ModalTrigger modal={<ServiceLBModalView serviceLB={ serviceLB }/>}>
					<tr key={ serviceLB.key } className="info">
						
						      
					</tr>
				</ModalTrigger>
			);
		});

		return (
        <div>
			<Table hover>
				<thead>
					<tr>
					
					      
					</tr>
				</thead>
				<tbody>
            		{ serviceLBListView }
				</tbody>
			</Table>
        </div>
    	);
	}
});

var ServiceLBModalView = React.createClass({
	render() {
		var obj = this.props.serviceLB
	    return (
	      <Modal {...this.props} bsStyle='primary' bsSize='large' title='ServiceLB' animation={false}>
	        <div className='modal-body' style={ {margin: '5%',} }>
			
			
				<Input type='text' label='Service ip' ref='ipAddress' defaultValue={obj.ipAddress} placeholder='Service ip' />
			
				<Input type='text' label='Service network name' ref='networkName' defaultValue={obj.networkName} placeholder='Service network name' />
			
				<Input type='text' label='service provider port' ref='ports' defaultValue={obj.ports} placeholder='service provider port' />
			
				<Input type='text' label='labels key value pair' ref='selectors' defaultValue={obj.selectors} placeholder='labels key value pair' />
			
				<Input type='text' label='service name' ref='serviceName' defaultValue={obj.serviceName} placeholder='service name' />
			
				<Input type='text' label='Tenant Name' ref='tenantName' defaultValue={obj.tenantName} placeholder='Tenant Name' />
			
			</div>
	        <div className='modal-footer'>
				<Button onClick={this.props.onRequestHide}>Close</Button>
	        </div>
	      </Modal>
	    );
  	}
});

module.exports.ServiceLBSummaryView = ServiceLBSummaryView
module.exports.ServiceLBModalView = ServiceLBModalView
var TenantSummaryView = React.createClass({
  	render: function() {
		var self = this

		// Walk thru all objects
		var tenantListView = self.props.tenants.map(function(tenant){
			return (
				<ModalTrigger modal={<TenantModalView tenant={ tenant }/>}>
					<tr key={ tenant.key } className="info">
						
						  
					</tr>
				</ModalTrigger>
			);
		});

		return (
        <div>
			<Table hover>
				<thead>
					<tr>
					
					  
					</tr>
				</thead>
				<tbody>
            		{ tenantListView }
				</tbody>
			</Table>
        </div>
    	);
	}
});

var TenantModalView = React.createClass({
	render() {
		var obj = this.props.tenant
	    return (
	      <Modal {...this.props} bsStyle='primary' bsSize='large' title='Tenant' animation={false}>
	        <div className='modal-body' style={ {margin: '5%',} }>
			
			
				<Input type='text' label='Network name' ref='defaultNetwork' defaultValue={obj.defaultNetwork} placeholder='Network name' />
			
				<Input type='text' label='Tenant Name' ref='tenantName' defaultValue={obj.tenantName} placeholder='Tenant Name' />
			
			</div>
	        <div className='modal-footer'>
				<Button onClick={this.props.onRequestHide}>Close</Button>
	        </div>
	      </Modal>
	    );
  	}
});

module.exports.TenantSummaryView = TenantSummaryView
module.exports.TenantModalView = TenantModalView
var VolumeSummaryView = React.createClass({
  	render: function() {
		var self = this

		// Walk thru all objects
		var volumeListView = self.props.volumes.map(function(volume){
			return (
				<ModalTrigger modal={<VolumeModalView volume={ volume }/>}>
					<tr key={ volume.key } className="info">
						
						      
					</tr>
				</ModalTrigger>
			);
		});

		return (
        <div>
			<Table hover>
				<thead>
					<tr>
					
					      
					</tr>
				</thead>
				<tbody>
            		{ volumeListView }
				</tbody>
			</Table>
        </div>
    	);
	}
});

var VolumeModalView = React.createClass({
	render() {
		var obj = this.props.volume
	    return (
	      <Modal {...this.props} bsStyle='primary' bsSize='large' title='Volume' animation={false}>
	        <div className='modal-body' style={ {margin: '5%',} }>
			
			
				<Input type='text' label='' ref='datastoreType' defaultValue={obj.datastoreType} placeholder='' />
			
				<Input type='text' label='' ref='mountPoint' defaultValue={obj.mountPoint} placeholder='' />
			
				<Input type='text' label='' ref='poolName' defaultValue={obj.poolName} placeholder='' />
			
				<Input type='text' label='' ref='size' defaultValue={obj.size} placeholder='' />
			
				<Input type='text' label='Tenant Name' ref='tenantName' defaultValue={obj.tenantName} placeholder='Tenant Name' />
			
				<Input type='text' label='Volume Name' ref='volumeName' defaultValue={obj.volumeName} placeholder='Volume Name' />
			
			</div>
	        <div className='modal-footer'>
				<Button onClick={this.props.onRequestHide}>Close</Button>
	        </div>
	      </Modal>
	    );
  	}
});

module.exports.VolumeSummaryView = VolumeSummaryView
module.exports.VolumeModalView = VolumeModalView
var VolumeProfileSummaryView = React.createClass({
  	render: function() {
		var self = this

		// Walk thru all objects
		var volumeProfileListView = self.props.volumeProfiles.map(function(volumeProfile){
			return (
				<ModalTrigger modal={<VolumeProfileModalView volumeProfile={ volumeProfile }/>}>
					<tr key={ volumeProfile.key } className="info">
						
						      
					</tr>
				</ModalTrigger>
			);
		});

		return (
        <div>
			<Table hover>
				<thead>
					<tr>
					
					      
					</tr>
				</thead>
				<tbody>
            		{ volumeProfileListView }
				</tbody>
			</Table>
        </div>
    	);
	}
});

var VolumeProfileModalView = React.createClass({
	render() {
		var obj = this.props.volumeProfile
	    return (
	      <Modal {...this.props} bsStyle='primary' bsSize='large' title='VolumeProfile' animation={false}>
	        <div className='modal-body' style={ {margin: '5%',} }>
			
			
				<Input type='text' label='' ref='datastoreType' defaultValue={obj.datastoreType} placeholder='' />
			
				<Input type='text' label='' ref='mountPoint' defaultValue={obj.mountPoint} placeholder='' />
			
				<Input type='text' label='' ref='poolName' defaultValue={obj.poolName} placeholder='' />
			
				<Input type='text' label='' ref='size' defaultValue={obj.size} placeholder='' />
			
				<Input type='text' label='Tenant Name' ref='tenantName' defaultValue={obj.tenantName} placeholder='Tenant Name' />
			
				<Input type='text' label='Volume profile Name' ref='volumeProfileName' defaultValue={obj.volumeProfileName} placeholder='Volume profile Name' />
			
			</div>
	        <div className='modal-footer'>
				<Button onClick={this.props.onRequestHide}>Close</Button>
	        </div>
	      </Modal>
	    );
  	}
});

module.exports.VolumeProfileSummaryView = VolumeProfileSummaryView
module.exports.VolumeProfileModalView = VolumeProfileModalView
