// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	oauth_v1 "github.com/openshift/api/oauth/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeOAuthClients implements OAuthClientInterface
type FakeOAuthClients struct {
	Fake *FakeOauthV1
}

var oauthclientsResource = schema.GroupVersionResource{Group: "oauth.openshift.io", Version: "v1", Resource: "oauthclients"}

var oauthclientsKind = schema.GroupVersionKind{Group: "oauth.openshift.io", Version: "v1", Kind: "OAuthClient"}

// Get takes name of the oAuthClient, and returns the corresponding oAuthClient object, and an error if there is any.
func (c *FakeOAuthClients) Get(name string, options v1.GetOptions) (result *oauth_v1.OAuthClient, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(oauthclientsResource, name), &oauth_v1.OAuthClient{})
	if obj == nil {
		return nil, err
	}
	return obj.(*oauth_v1.OAuthClient), err
}

// List takes label and field selectors, and returns the list of OAuthClients that match those selectors.
func (c *FakeOAuthClients) List(opts v1.ListOptions) (result *oauth_v1.OAuthClientList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(oauthclientsResource, oauthclientsKind, opts), &oauth_v1.OAuthClientList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &oauth_v1.OAuthClientList{ListMeta: obj.(*oauth_v1.OAuthClientList).ListMeta}
	for _, item := range obj.(*oauth_v1.OAuthClientList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested oAuthClients.
func (c *FakeOAuthClients) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(oauthclientsResource, opts))
}

// Create takes the representation of a oAuthClient and creates it.  Returns the server's representation of the oAuthClient, and an error, if there is any.
func (c *FakeOAuthClients) Create(oAuthClient *oauth_v1.OAuthClient) (result *oauth_v1.OAuthClient, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(oauthclientsResource, oAuthClient), &oauth_v1.OAuthClient{})
	if obj == nil {
		return nil, err
	}
	return obj.(*oauth_v1.OAuthClient), err
}

// Update takes the representation of a oAuthClient and updates it. Returns the server's representation of the oAuthClient, and an error, if there is any.
func (c *FakeOAuthClients) Update(oAuthClient *oauth_v1.OAuthClient) (result *oauth_v1.OAuthClient, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(oauthclientsResource, oAuthClient), &oauth_v1.OAuthClient{})
	if obj == nil {
		return nil, err
	}
	return obj.(*oauth_v1.OAuthClient), err
}

// Delete takes name of the oAuthClient and deletes it. Returns an error if one occurs.
func (c *FakeOAuthClients) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(oauthclientsResource, name), &oauth_v1.OAuthClient{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeOAuthClients) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(oauthclientsResource, listOptions)

	_, err := c.Fake.Invokes(action, &oauth_v1.OAuthClientList{})
	return err
}

// Patch applies the patch and returns the patched oAuthClient.
func (c *FakeOAuthClients) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *oauth_v1.OAuthClient, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(oauthclientsResource, name, data, subresources...), &oauth_v1.OAuthClient{})
	if obj == nil {
		return nil, err
	}
	return obj.(*oauth_v1.OAuthClient), err
}
