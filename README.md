# But 

Le but de ce projet est de montrer qu'on peut utiliser metarank pour pouvoir proposer aux utilisateurs une selections d'annonces similaire par rapport à ce qu'il cherche.

## Prérequis

Avant de commencer, assurez-vous d'avoir installé les outils suivants sur votre machine :

- [Go](https://golang.org/doc/install) (version 1.16 ou supérieure)
- [Docker](https://www.docker.com/get-started) pour exécuter Metarank
- [Metarank](https://github.com/metarank/metarank) (pour l'entraînement et le déploiement des modèles)

## How to use it

### Étape 1 : Cloner le dépôt

- Clonez le dépôt sur votre machine locale

``` bash 
git clone git@github.com:AdeleRICHARD/metarankPoc.git
cd metarankPoc
```

## Étape 2 : Configurer Metarank

### Préparer les données :

Placez vos données d'annonces dans un fichier JSONL dans le dossier `data/`.
Utilisez le script `main.go` pour créer des événements utilisateur et interaction à partir du json d'annonces. 
Il faut au moins 100 annonces différentes. 

### Configurer Metarank :

Le fichier `config_similarity.yml` est celui qui fait la configuration pour un métarank qui peut donner des similarités ou des trendings sur les actions des utilisateurs. Ce qui n'est pas ce que nous voulons actuellement.

Un deuxème fichier de `config.yml` a été créée et utilisé pour créer un metarank capable de nous donner les recommandations basées sur l'id des annonces.

### Pour un modèle de similarité seulement 

Un fichier de config a déjà été généré dans ce projet. Mais si vous voulez changer de config en fonction de vos données, vous pouvez utiliser l'autofeature de metarank. ATTENTION, l'autofeature fonctionne seulement pour les modèles avec intéraction utilisateur.

Exemple : 
```bash
docker run -v $(pwd)/user_interaction_events_100.jsonl:/input/events.jsonl -v $(pwd):/output metarank/metarank autofeature --data /input/events.jsonl --out /output/config.yml
```
Les données dans le jsonl doivent être triées et doivent comporter au moins 10 clicks utilisateurs et au moins 100 events. 
Il faut aussi que ça soit sorted par timestamp, mais metarank a une commande de sort.

Modifiez le fichier `config.yml` pour définir vos modèles de recommandation, tels que `similar`, `trending`, etc.

### Démarrer Metarank avec Docker :

Utilisez Docker pour lancer Metarank soit en mode serveur, donc non entrainé :

```bash
docker run -p 8080:8080 -v $(pwd)/config.yml:/config.yml -v $(pwd)/data:/data metarank/metarank serve --config /config.yml
```

Soit directement en api entrainée : 
*ATTENTION* quand on fait la commande suivante, si on a le vpn l'appel peut ne pas passer car il appel l'api de huggingface et j'ai l'impression que ça fait conflit.

```bash
docker run -i -t -p 8080:8080 -v $(pwd):/opt/metarank metarank/metarank:latest standalone --config /opt/metarank/config.yml --data /opt/metarank/formatted_classifieds.jsonl
```
Cette deuxième approche permet une première initialisation de métarank avec nos bonnes données et notre bonne config pour une service prêt à l'emploi. Avec un set de données à 100 cela prend à peine quelques secondes, mais avec plus de 100 000 annonces nous sommes à 12 minutes.


On pourra ensuite l'entraîner via l'appel api [feedback](https://docs.metarank.ai/reference/api#feedback) : 
```bash
curl -X POST http://localhost:8080/feedback -H "Content-Type: application/json" -d '[
  {
    "event": "item",
    "id": "65929736",
    "timestamp": "1723647452",
    "item": "65929736",
    "fields": [
      {"name": "price", "value": 329000},
      {"name": "estateType", "value": "maison"},
      {"name": "city", "value": "Gagny (93)"},
      {"name": "postalCode", "value": "93220"},
      {"name": "transaction", "value": "vente"}
    ]
  },
  {
    "event": "item",
    "id": "40272755",
    "timestamp": "1723647453",
    "item": "40272755",
    "fields": [
      {"name": "price", "value": 329000},
      {"name": "estateType", "value": "maison"},
      {"name": "city", "value": "Gagny (93)"},
      {"name": "postalCode", "value": "93220"},
      {"name": "transaction", "value": "vente"}
    ]
  },
  {
    "event": "item",
    "id": "65082628",
    "timestamp": "1723647454",
    "item": "65082628",
    "fields": [
      {"name": "price", "value": 371000},
      {"name": "estateType", "value": "maison"},
      {"name": "city", "value": "Gagny (93)"},
      {"name": "postalCode", "value": "93220"},
      {"name": "transaction", "value": "vente"}
    ]
  },
  {
    "event": "item",
    "id": "67481504",
    "timestamp": "1723647455",
    "item": "67481504",
    "fields": [
      {"name": "price", "value": 750000},
      {"name": "estateType", "value": "maison"},
      {"name": "city", "value": "Gagny (93)"},
      {"name": "postalCode", "value": "93220"},
      {"name": "transaction", "value": "vente"}
    ]
  },
  {
    "event": "item",
    "id": "58450982",
    "timestamp": "1723647456",
    "item": "58450982",
    "fields": [
      {"name": "price", "value": 349000},
      {"name": "estateType", "value": "maison"},
      {"name": "city", "value": "Gagny (93)"},
      {"name": "postalCode", "value": "93220"},
      {"name": "transaction", "value": "vente"}
    ]
  }
]'
```

Plus tard on peut même ajouter le type user pour lier nos annonces à des utilisateurs.

## Tester le programme et l'api

~~Pour tester que notre modèle est entraîné correctement et qu'on a les appels api qui fonctionnent, il suffit de lancer la commande de test :~~

~~```bash~~
~~go test -v~~
~~```~~


Ou alors de faire un curl pour avoir les recommendations de notre modèle par rapport à une annonce : 

```bash
curl -X POST http://localhost:8080/recommend/semantic -H "Content-Type: application/json" -d '{
  "count":5,
"items":["67492332"]
}'
```
Ce que nous donne l'api / modèle après un entraînement sur 100 000 annonces :

```json
[
    {
        "event": "item",
        "id": "68726616",
        "timestamp": "1723821685",
        "item": "68726616",
        "fields": [
            {"name": "price", "value": 279000},
            {"name": "estateType", "value": "appartement"},
            {"name": "city", "value": "Cagnes-sur-Mer (06)"},
            {"name": "postalCode", "value": "06800"},
            {"name": "transaction", "value": "vente"}
        ]
    },
    {
        "event": "item",
        "id": "70996076",
        "timestamp": "1723821685",
        "item": "70996076",
        "fields": [
            {"name": "price", "value": 510000},
            {"name": "estateType", "value": "appartement"},
            {"name": "city", "value": "Cagnes-sur-Mer (06)"},
            {"name": "postalCode", "value": "06800"},
            {"name": "transaction", "value": "vente"}
        ]
    },
    {
        "event": "item",
        "id": "69884780",
        "timestamp": "1723821685",
        "item": "69884780",
        "fields": [
            {"name": "price", "value": 275000},
            {"name": "estateType", "value": "appartement"},
            {"name": "city", "value": "Cagnes-sur-Mer (06)"},
            {"name": "postalCode", "value": "06800"},
            {"name": "transaction", "value": "vente"}
        ]
    },
    {
        "event": "item",
        "id": "71709762",
        "timestamp": "1723821685",
        "item": "71709762",
        "fields": [
            {"name": "price", "value": 343000},
            {"name": "estateType", "value": "appartement"},
            {"name": "city", "value": "Cagnes-sur-Mer (06)"},
            {"name": "postalCode", "value": "06800"},
            {"name": "transaction", "value": "vente"}
        ]
    },
    {
        "event": "item",
        "id": "69505692",
        "timestamp": "1723821685",
        "item": "69505692",
        "fields": [
            {"name": "price", "value": 285000},
            {"name": "estateType", "value": "appartement"},
            {"name": "city", "value": "Cagnes-sur-Mer (06)"},
            {"name": "postalCode", "value": "06800"},
            {"name": "transaction", "value": "vente"}
        ]
    }
]

```

# L'intégration
[doc](https://docs.metarank.ai/reference/deployment-overview/kubernetes)
## Prérequis
Il faudrait déjà créer le container métarank et lui associer un redis.
En terme de ressources que cela utiliserais je ne suis pas sûr. 

---

**pros**
- Une solution avec intégration d'IA et qui peut être amélioré plus facilement par la suite. On pourrait vouloir ajouter un solution de ranking éventuellement. 
- Des résultats qui semblent cohérents sans avoir besoin de trop toucher à une histoire de poids ou autre.
- Une manière plutôt simple d'entraîner le modèle.

**cons**
- Metarank n'a pas de stockage de mémoire long term. Cela veut dire qu'à tout moment si le pod redémarre, on perd tout. Il lui faut forcément un redis pour stocker les valeurs.
- On se pose la question de la rotation des annonces. A combien de temps on estime la durée du stockage.
- Gestion des résultats qui peut demander à être plus fine que si c'est ES qui nous donne les résultats.
- Un service en plus à aller chercher.
