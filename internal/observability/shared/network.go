package shared

import (
	"context"
	"errors"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

func EnsureNetwork(ctx context.Context, cli *client.Client) error {
	args := filters.NewArgs()
	args.Add("name", ObservabilityNetworkName)
	args.Add("name", L1NetworkName)
	args.Add("name", L2NetworkName)

	networks, err := cli.NetworkList(ctx, network.ListOptions{Filters: args})
	if err != nil {
		return errors.Join(err, errors.New("failed to list Docker network"))
	}

	var l1Available, l2Available, observabilityAvailable bool
	for _, network := range networks {
		if network.Name == L1NetworkName {
			l1Available = true
		}
		if network.Name == L2NetworkName {
			l2Available = true
		}
		if network.Name == ObservabilityNetworkName {
			observabilityAvailable = true
		}
	}

	if !l1Available {
		_, err = cli.NetworkCreate(ctx, L1NetworkName, network.CreateOptions{
			Driver: "bridge",
			Labels: Labels,
		})
		if err != nil {
			return errors.Join(err, errors.New("failed to create L1 network"))
		}
	}

	if !l2Available {
		_, err = cli.NetworkCreate(ctx, L2NetworkName, network.CreateOptions{
			Driver: "bridge",
			Labels: Labels,
		})
		if err != nil {
			return errors.Join(err, errors.New("failed to create L2 network"))
		}
	}

	if !observabilityAvailable {
		_, err = cli.NetworkCreate(ctx, ObservabilityNetworkName, network.CreateOptions{
			Driver: "bridge",
			Labels: Labels,
		})
		if err != nil {
			return errors.Join(err, errors.New("failed to create observability network"))
		}
	}

	return nil
}

// getAvailableLocalnetNetworks returns list of localnet networks that exist
func getAvailableLocalnetNetworks(ctx context.Context, cli *client.Client) ([]string, error) {
	args := filters.NewArgs()
	args.Add("name", L1NetworkName)
	args.Add("name", L2NetworkName)

	networks, err := cli.NetworkList(ctx, network.ListOptions{Filters: args})
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to list Docker networks"))
	}

	var available []string
	for _, net := range networks {
		if net.Name == L1NetworkName || net.Name == L2NetworkName {
			available = append(available, net.Name)
		}
	}

	return available, nil
}

// AttachToLocalnetNetworks attaches a container to all available localnet networks
func AttachToLocalnetNetworks(ctx context.Context, cli *client.Client, containerID string) error {
	localnetNetworks, err := getAvailableLocalnetNetworks(ctx, cli)
	if err != nil {
		return err
	}

	for _, networkName := range localnetNetworks {
		if err := cli.NetworkConnect(ctx, networkName, containerID, nil); err != nil {
			return errors.Join(err, errors.New("failed to connect container to "+networkName))
		}
	}

	return nil
}
