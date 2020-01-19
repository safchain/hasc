import React, { Component } from 'react';
import { ScrollView, View, TouchableOpacity, Text, TextInput, Button } from 'react-native';
import { createAppContainer, NavigationEvents } from 'react-navigation';
import { createDrawerNavigator } from 'react-navigation-drawer';
import { createStackNavigator } from 'react-navigation-stack';
import Icon from 'react-native-vector-icons/FontAwesome';
import AsyncStorage from '@react-native-community/async-storage';
import { WebView } from 'react-native-webview';
import { NetworkInfo } from 'react-native-network-info';

class Main extends Component {

  constructor(props) {
    super(props);

    this.state = {
      url: ''
    }
  }

  applySettings = async () => {
    try {
      const value = await AsyncStorage.getItem('@settings');
      if (value !== null) {
        const data = JSON.parse(value);

        NetworkInfo.getSSID().then(ssid => {
          if (ssid === data.wifiSSID) {
            const url = `http://${data.wifiAddress}/?mode=native&username=${data.username}&password=${data.password}`
            this.setState({ url: url });
          } else {
            const url = `http://${data.remoteAddress}/?mode=native&username=${data.username}&password=${data.password}`
            this.setState({ url: url });
          }
        });
      }
    } catch (e) {
      console.log(e)
    }
  }

  render() {
    return (
      <>
        <NavigationEvents
          onDidFocus={() => this.applySettings()}
        />
        <WebView
          source={{ uri: this.state.url }}
        />
      </>
    );
  }
}

class Settings extends Component {

  constructor(props) {
    super(props);

    this.state = {
      wifiAddress: '',
      wifiSSID: '',
      remoteAddress: '',
      username: '',
      password: ''
    };
  }

  loadSettings = async () => {
    try {
      const value = await AsyncStorage.getItem('@settings');
      if (value !== null) {
        const data = JSON.parse(value);
        this.setState(data);
      }
    } catch (e) {
      console.log(e)
    }
  }

  storeSettings = async () => {
    try {
      await AsyncStorage.setItem('@settings', JSON.stringify(this.state))
    } catch (e) {
      console.log(e)
    }
  }

  updateSettings = function (key, value) {
    var entry = {};
    entry[key] = value;

    this.setState(entry, this.storeSettings);
  }

  setSSID = function () {
    NetworkInfo.getSSID().then(ssid => {
      this.setState({ "wifiSSID": ssid }, this.storeSettings);
    });
  }

  render() {
    return (
      <ScrollView style={{ margin: 10 }}>
        <NavigationEvents
          onDidFocus={() => this.loadSettings()}
        />
        <Text style={{ fontSize: 18 }}>Wifi address :</Text>
        <TextInput
          placeholder="192.168.0.1:12345"
          placeholderTextColor="#888"
          defaultValue={this.state.wifiAddress}
          onChangeText={this.updateSettings.bind(this, 'wifiAddress')}
        />
        <Text style={{ fontSize: 18 }}>Wifi SSID :</Text>
        <TextInput
          placeholder="192.168.0.1:12345"
          placeholderTextColor="#888"
          defaultValue={this.state.wifiSSID}
          onChangeText={this.updateSettings.bind(this, 'wifiSSID')}
        />
        <View style={{ marginBottom: 15 }}>
          <Button
            title="Current SSID"
            onPress={this.setSSID.bind(this)}
          />
        </View>
        <Text style={{ fontSize: 18 }}>Remote address :</Text>
        <TextInput
          placeholder="192.168.0.1:12345"
          placeholderTextColor="#888"
          defaultValue={this.state.remoteAddress}
          onChangeText={this.updateSettings.bind(this, 'remoteAddress')}
        />
        <Text style={{ fontSize: 18 }}>Username :</Text>
        <TextInput
          placeholder="admin"
          placeholderTextColor="#888"
          defaultValue={this.state.username}
          onChangeText={this.updateSettings.bind(this, 'username')}
        />
        <Text style={{ fontSize: 18 }}>Password :</Text>
        <TextInput
          placeholder="admin"
          placeholderTextColor="#888"
          secureTextEntry={true}
          defaultValue={this.state.password}
          onChangeText={this.updateSettings.bind(this, 'password')}
        />
      </ScrollView>
    );
  }
}

class NavigationDrawer extends Component {
  toggleDrawer = () => {
    this.props.navigationProps.toggleDrawer();
  };
  render() {
    return (
      <View style={{ flexDirection: 'row' }}>
        <TouchableOpacity onPress={this.toggleDrawer.bind(this)}>
          <Icon name="bars" size={24} color="#fff" style={{ marginLeft: 20 }} />
        </TouchableOpacity>
      </View>
    );
  }
}

const MainStackNavigator = createStackNavigator({
  First: {
    screen: Main,
    navigationOptions: ({ navigation }) => ({
      title: 'H.A.S.C.',
      headerLeft: () => <NavigationDrawer navigationProps={navigation} />,
      headerStyle: {
        backgroundColor: '#3f51b5',
      },
      headerTintColor: '#fff',
    }),
  },
});

const SettingsStackNavigator = createStackNavigator({
  Second: {
    screen: Settings,
    navigationOptions: ({ navigation }) => ({
      title: 'H.A.S.C.',
      headerLeft: () => <NavigationDrawer navigationProps={navigation} />,
      headerStyle: {
        backgroundColor: '#3f51b5',
      },
      headerTintColor: '#fff',
    }),
  },
});

const DrawerNavigator = createDrawerNavigator({
  Main: {
    screen: MainStackNavigator,
    navigationOptions: {
      drawerLabel: 'Home',
    },
  },
  Settings: {
    screen: SettingsStackNavigator,
    navigationOptions: {
      drawerLabel: 'Settings',
    },
  },
}, {
  drawerType: "back",
  initialRouteName: "Main"
});

export default createAppContainer(DrawerNavigator);