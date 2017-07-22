import React from 'react';
import {
  AppRegistry,
  Text,
  ScrollView,
  TextInput,
  AsyncStorage,
  StyleSheet,
  TouchableOpacity
} from 'react-native';
import { DrawerNavigator, StackNavigator } from 'react-navigation';
import { Button, View, WebView, TouchableHighlight } from 'react-native';
import Icon from 'react-native-vector-icons/FontAwesome';

const styles = StyleSheet.create({
  welcome: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center'
  },
  welcomeTitle:{
    fontWeight: 'bold',
    color: '#163',
    textAlign: 'center',
    fontSize: 20,
  },
  welcomeButton: {
    flex: 0.1,
    flexDirection: 'row',
    alignItems: 'flex-end',
    justifyContent: 'flex-end',
    marginRight: 30,
    marginBottom: 30
  }
});

class Home extends React.Component {

  static navigationOptions = function(props) {
    return {
      drawerLabel: 'Home',
      drawerIcon: <Icon name="home" size={24} color="#111" />
    }
  }

  constructor(props) {
    super(props);
    this.state = {
      'source': {
        'uri': ''
      }
    };
  }

  componentDidMount() {
    AsyncStorage.getItem('wifiAddress').then((value) => {
      var uri = 'http://' + value + '/?header=false';
      fetch(uri).then(() =>
        this.setState({ 'source': {
            'uri': uri
          }
        })
      )
      .catch((error) => {
        this.setState({ 'source': {
            'uri': ''
          }
        })
      });
    });
  }

  render() {
    if (this.state.source.uri !== '') {
      return (
        <WebView
          source={this.state.source}
        />
      )
    } else {
      return (
        <View style={{flex:1}}>
          <View style={styles.welcome}>
            <Text style={styles.welcomeTitle}>Welcome to H.A.S.C.</Text>
          </View>
          <View style={styles.welcomeButton}>
            <TouchableOpacity
              style={{
                  borderWidth:1,
                  borderColor:'rgba(0,0,0,0.2)',
                  alignItems:'center',
                  justifyContent:'center',
                  width:60,
                  height:60,
                  backgroundColor:'#004d00',
                  borderRadius:100,
                }}

              onPress={() => {
                  this.props.navigation.navigate('Settings');
                }
              }
            >
              <Icon name={"chevron-right"}  size={20} color="#fff" />
            </TouchableOpacity>
          </View>
        </View>
      )
    }
  }
}

class Settings extends React.Component {

  constructor(props) {
    super(props);
    this.state = { 'wifiAddress': '' };
  }

  componentDidMount() {
    AsyncStorage.getItem('wifiAddress').then((value) => {
      this.setState({ 'wifiAddress': value });
    })
  }

  setWifiAddress = function(value) {
    AsyncStorage.setItem('wifiAddress', value);
  }

  static navigationOptions = function(props) {
    return {
      drawerLabel: 'Settings',
      drawerIcon:  <Icon name="gear" size={24} color="#111" />
    }
  }

  render() {
    return (
      <ScrollView style={{margin: 10}}>
        <Text style={{fontSize:18}}>Wifi address :</Text>
        <TextInput
          placeholder="192.168.0.1:12345"
          placeholderTextColor="#888"
          onChangeText={this.setWifiAddress}
          defaultValue={this.state.wifiAddress}/>
        <Text style={{fontSize:18}}>Wifi SSID :</Text>
        <TextInput
          placeholder="192.168.0.1:12345"
          placeholderTextColor="#888"/>
        <Text style={{fontSize:18}}>Remote address :</Text>
        <TextInput
          placeholder="192.168.0.1:12345"
          placeholderTextColor="#888"/>
        <Text style={{fontSize:18}}>Username :</Text>
        <TextInput
          placeholder="admin"
          placeholderTextColor="#888"/>
        <Text style={{fontSize:18}}>Password :</Text>
        <TextInput
          placeholder="admin"
          placeholderTextColor="#888"
          secureTextEntry={true}/>
      </ScrollView>
    );
  }
}

const Drawer = DrawerNavigator({
  Home: { screen: Home },
  Settings: { screen: Settings}
});

Drawer.navigationOptions = ({ navigation }) => ({
  title: 'H.A.S.C',
  headerStyle: {
    backgroundColor: '#004d00',
  },
  headerTitleStyle: {
    color: 'white',
  },
  headerTintColor: 'white',
  headerLeft:
    <View style={{marginLeft: 10}}>
      <Icon name="bars" size={24} color="#eee" onPress={() => {
        if (navigation.state.index === 0) {
          navigation.navigate('DrawerOpen');
        } else {
          navigation.navigate('DrawerClose');
        }}
      }/>
    </View>
});

const App = StackNavigator({
  Drawer: { screen: Drawer }
});

AppRegistry.registerComponent('app', () => App);
